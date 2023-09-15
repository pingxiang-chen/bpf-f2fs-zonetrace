package server

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/respbuffer"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/server/statics"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

// api represents the API handlers for the server.
type api struct {
	znsMemory znsmemory.ZNSMemory
}

// indexHandler handles the root URL, redirecting to "/zone/0".
func (s *api) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/zone/0", http.StatusFound)
}

// htmlHandler serves HTML content.
func (s *api) htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write(statics.IndexHtmlFile)
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

// staticsHandler serves static files.\
func (s *api) staticsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse /static/:filename
	staticFileName := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]
	staticFile, ok := statics.StaticFileMap[staticFileName]
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

// zoneInfoHandler handles requests for zone information.
func (s *api) zoneInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse /api/info/:pk
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

// getMostFrequentSegmentType finds the most frequent segment type
func getMostFrequentSegmentType(segments []Segment) znsmemory.SegmentType {
	countSegmentType := make(map[znsmemory.SegmentType]int)
	for _, segment := range segments {
		countSegmentType[znsmemory.SegmentType(segment.SegmentType)]++
	}

	mostFrequentSegmentType := znsmemory.NotChanged
	maxCount := 0
	for segmentType, count := range countSegmentType {
		if segmentType == znsmemory.NotChanged || segmentType == znsmemory.UnknownSegment {
			continue
		}
		if count > maxCount {
			mostFrequentSegmentType = segmentType
			maxCount = count
		}
	}
	return mostFrequentSegmentType
}

// streamZoneDataHandler handles streaming zone data to clients.
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

	// send initial zone segment type
	info := s.znsMemory.GetZoneInfo()
	for i := 0; i < info.TotalZone; i++ {
		zone, err := s.znsMemory.GetZone(i)
		if err != nil {
			http.Error(w, "Error getting zone", http.StatusInternalServerError)
			return
		}
		data := ToZoneResponse(zone.ZoneNo, zone.LastSegmentType, nil)
		if _, err := w.Write(data.Serialize()); err != nil {
			http.Error(w, "Error writing data", http.StatusInternalServerError)
			return
		}
	}

	// send initial segments data
	zone, err := s.znsMemory.GetZone(currentZoneNo)
	if err != nil {
		http.Error(w, "Error getting zone", http.StatusInternalServerError)
		return
	}
	segments := make([]Segment, len(zone.Segments))
	for i, segment := range zone.Segments {
		segments[i] = Segment{
			SegmentNo:   i,
			SegmentType: int(segment.SegmentType),
			Map:         segment.ValidMap,
		}
	}
	mostFrequentSegmentType := getMostFrequentSegmentType(segments)
	data := ToZoneResponse(zone.ZoneNo, mostFrequentSegmentType, segments)
	respBuf.Push(data.Serialize())

	// subscribe zone updates
	sub := s.znsMemory.Subscribe()
	defer s.znsMemory.UnSubscribe(sub)
	needUpdateSegment := make(map[int]struct{})
	lastZoneSegmentType := make(map[int]znsmemory.SegmentType)

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
			segmentType := update.SegmentType
			if update.ZoneDirtyCount == 0 {
				segmentType = znsmemory.EmptySegment
			}
			fmt.Println("update", update.ZoneNo, update.SegmentNo, segmentType)

			// for update segment type
			previousSegmentType, ok := lastZoneSegmentType[update.ZoneNo]
			if ok {
				if previousSegmentType == segmentType {
					continue
				}
			}
			lastZoneSegmentType[update.ZoneNo] = segmentType

			// notice only zone number
			data = ToZoneResponse(update.ZoneNo, segmentType, nil)
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
					SegmentNo:   segmentNo,
					Map:         seg.ValidMap,
					SegmentType: int(seg.SegmentType),
				})
				delete(needUpdateSegment, segmentNo)
			}
			mostFrequentSegmentType = getMostFrequentSegmentType(segments)
			data = ToZoneResponse(currentZoneNo, mostFrequentSegmentType, segments)
			respBuf.Push(data.Serialize())
		}
	}
}

// installGracefulShutdown installs a graceful shutdown for the HTTP server.
// It will wait for the remain requests to be done and then shutdown the server.
func installGracefulShutdown(ctx context.Context, server *http.Server) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fmt.Println("Shutdown server...")
		_ = server.Shutdown(shutdownCtx)
	}()
}

// New creates a new instance of the HTTP server.
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
