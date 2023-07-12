package server

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/respbuffer"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed statics/index.html
var indexHtml []byte

type api struct {
	znsMemory znsmemory.ZNSMemory
}

func (s *api) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/zone/0", http.StatusFound)
}

func (s *api) htmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write(indexHtml)
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

func (s *api) zoneInfoHandler(w http.ResponseWriter, r *http.Request) {
	zoneInfo := s.znsMemory.GetZoneInfo()
	data := ToZoneInfoResponse(*zoneInfo)
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(data.Serialize())
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

func (s *api) streamZoneDataHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("request panic:", err)
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

	// send all segments
	for i, segment := range zone.Segments {
		data := ToSegmentResponse(zone.ZoneNo, i, znsmemory.UnknownSegment, segment.ValidMap)
		if ok = sendSegment(w, flusher, data); !ok {
			return
		}
	}
	ctx, cancel := context.WithCancel(r.Context())
	respBuf := respbuffer.New(ctx)

	// write goroutine
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

	// subscribe zone updates
	sub := s.znsMemory.Subscribe()
	defer s.znsMemory.UnSubscribe(sub)
	lastUpdateZone := make(map[zoneNoSegmentTypePair]time.Time)
	needUpdateSegment := make(map[int]struct{})

	ticker := time.NewTicker(500 * time.Millisecond)
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
				// for same zoneNo
				needUpdateSegment[update.SegmentNo] = struct{}{}
				continue
			}
			// for different zoneNo
			now := time.Now()
			if update.ZoneNo != currentZoneNo {
				updatedZone := zoneNoSegmentTypePair{
					ZoneNo:      update.ZoneNo,
					SegmentType: update.SegmentType,
				}
				last := lastUpdateZone[updatedZone]
				// skip send updates if last update of same zoneNo is less than 500ms
				if now.Sub(last) < 500*time.Millisecond {
					continue
				}
				// last update is more than 500ms
				lastUpdateZone[updatedZone] = now
				// notice only zone number
				data := ToSegmentResponse(update.ZoneNo, 0, update.SegmentType, nil)
				respBuf.Push(data.Serialize())
				continue
			}
		case <-ticker.C:
			for segmentNo := range needUpdateSegment {
				segments, err := s.znsMemory.GetSegment(currentZoneNo, segmentNo)
				if err != nil {
					fmt.Println("Error getting segment", err)
					continue
				}
				data := ToSegmentResponse(currentZoneNo, segmentNo, znsmemory.UnknownSegment, segments.ValidMap)
				respBuf.Push(data.Serialize())
				delete(needUpdateSegment, segmentNo)
			}
		}
	}
}

func sendSegment(w http.ResponseWriter, flusher http.Flusher, resp SegmentResponse) bool {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		fmt.Println("Error serializing segment", err)
		return true
	}
	_, err := w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return false
	}
	flusher.Flush()
	return true
}

func installGracefulShutdown(ctx context.Context, server *http.Server) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fmt.Println("\nShutdown server...")
		_ = server.Shutdown(shutdownCtx)
	}()
}

func New(ctx context.Context, znsMemory znsmemory.ZNSMemory, port int) *http.Server {
	a := api{
		znsMemory: znsMemory,
	}
	handler := http.NewServeMux()
	handler.HandleFunc("/", a.indexHandler)
	handler.HandleFunc("/zone/", a.htmlHandler)
	handler.HandleFunc("/api/zone/info", a.zoneInfoHandler)
	handler.HandleFunc("/api/zone/", a.streamZoneDataHandler)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	installGracefulShutdown(ctx, server)
	return server
}
