package server

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/respbuffer"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/server/statics"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type api struct {
	znsMemory znsmemory.ZNSMemory
}

func (s *api) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/zone/0", http.StatusFound)
}

func (s *api) htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write(statics.IndexHtmlFile)
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

func (s *api) staticsHandler(w http.ResponseWriter, r *http.Request) {
	pk := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]
	staticFile, ok := statics.StaticFileMap[pk]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", staticFile.ContentType)
	_, err := w.Write(staticFile.File)
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

func (s *api) zoneInfoHandler(w http.ResponseWriter, r *http.Request) {
	pk := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]
	currentZoneNo, err := strconv.Atoi(pk)
	if err != nil {
		http.Error(w, "Invalid zone number", http.StatusBadRequest)
		return
	}
	zoneInfo := s.znsMemory.GetZoneInfo()
	zone, err := s.znsMemory.GetZone(currentZoneNo)
	if err != nil {
		http.Error(w, "Error getting zone", http.StatusInternalServerError)
		return
	}
	data := ToZoneInfoResponse(*zoneInfo, zone.LastSegmentType)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data.Serialize())
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

func (s *api) streamZoneDataHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("request panic:", err)
			debug.PrintStack()
		}
	}()
	pk := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]
	currentZoneNo, err := strconv.Atoi(pk)
	if err != nil {
		http.Error(w, "Invalid zone number", http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	zone, err := s.znsMemory.GetZone(currentZoneNo)
	if err != nil {
		http.Error(w, "Error getting zone", http.StatusInternalServerError)
		return
	}

	// write goroutine
	ctx, cancel := context.WithCancel(r.Context())
	respBuf := respbuffer.New(ctx)
	go func() {
		defer func() {
			// sometimes flusher.Flush() panics
			recover()
		}()
		for {
			data, err := respBuf.PopFirst()
			if err != nil {
				return
			}
			_, err = w.Write(data)
			if err != nil {
				http.Error(w, "Error writing data", http.StatusInternalServerError)
				cancel()
				return
			}
			flusher.Flush()
		}
	}()

	// send all segments
	segments := make([]Segment, len(zone.Segments))
	for i, segment := range zone.Segments {
		segments[i] = Segment{
			SegmentNo: i,
			Map:       segment.ValidMap,
		}
	}
	data := ToZoneResponse(zone.ZoneNo, znsmemory.NotChanged, segments)
	respBuf.Push(data.Serialize())

	// subscribe zone updates
	sub := s.znsMemory.Subscribe()
	defer s.znsMemory.UnSubscribe(sub)
	lastZoneUpdateTime := make(map[zoneNoSegmentTypePair]time.Time)
	needUpdateSegment := make(map[int]struct{})

	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-sub.Event:
			if update.ZoneNo == currentZoneNo {
				needUpdateSegment[update.SegmentNo] = struct{}{}
			}

			// for update segment type
			now := time.Now()
			updatedZone := zoneNoSegmentTypePair{ZoneNo: update.ZoneNo, SegmentType: update.SegmentType}
			last := lastZoneUpdateTime[updatedZone]
			// skip send updates if last update of same zoneNo is less than 500ms
			if now.Sub(last) < 500*time.Millisecond {
				continue
			}
			lastZoneUpdateTime[updatedZone] = now
			// notice only zone number
			data = ToZoneResponse(update.ZoneNo, update.SegmentType, nil)
			respBuf.Push(data.Serialize())
		case <-ticker.C:
			if len(needUpdateSegment) == 0 {
				continue
			}
			segments = make([]Segment, 0, len(needUpdateSegment))
			for segmentNo := range needUpdateSegment {
				seg, err := s.znsMemory.GetSegment(currentZoneNo, segmentNo)
				if err != nil {
					fmt.Println("Error getting segment", err)
					continue
				}
				segments = append(segments, Segment{
					SegmentNo: segmentNo,
					Map:       seg.ValidMap,
				})
				delete(needUpdateSegment, segmentNo)
			}
			data = ToZoneResponse(currentZoneNo, znsmemory.NotChanged, segments)
			respBuf.Push(data.Serialize())
		}
	}
}

func installGracefulShutdown(ctx context.Context, server *http.Server) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fmt.Println("Shutdown server...")
		_ = server.Shutdown(shutdownCtx)
	}()
}

func New(ctx context.Context, znsMemory znsmemory.ZNSMemory, port int) *http.Server {
	a := api{
		znsMemory: znsMemory,
	}
	handler := http.NewServeMux()
	handler.HandleFunc("/zone/", a.htmlHandler)
	handler.HandleFunc("/api/info/", a.zoneInfoHandler)
	handler.HandleFunc("/api/zone/", a.streamZoneDataHandler)
	handler.HandleFunc("/static/", a.staticsHandler)
	handler.HandleFunc("/", a.indexHandler)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	installGracefulShutdown(ctx, server)
	return server
}
