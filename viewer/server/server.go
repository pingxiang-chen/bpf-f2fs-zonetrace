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

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/fstool"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/respbuffer"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/static"
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
	_, err := w.Write(static.IndexHtmlFile)
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
}

// staticsHandler serves static files.\
func (s *api) staticsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse /static/:filename
	staticFileName := r.RequestURI[strings.Index(r.RequestURI, "/static/")+8:]
	staticFile, ok := static.ServingFileMap[staticFileName]
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

// highlightHandler handles the highlight URL, redirecting to "/highlight/0".
func (s *api) highlightHandler(w http.ResponseWriter, r *http.Request) {
	pk := r.RequestURI[strings.LastIndex(r.RequestURI, "/")+1:]
	if pk == "" {
		http.Redirect(w, r, "/highlight/0", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write(static.HighlightHtmlFile)
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

	// make new goroutine for write data
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
		data := ToZoneResponse(zone.ZoneNo, zone.FrequentSegmentType(), nil)
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
	data := ToZoneResponse(zone.ZoneNo, zone.FrequentSegmentType(), segments)
	respBuf.Push(data.Serialize())

	// subscribe zone updates
	sub := s.znsMemory.Subscribe()
	defer s.znsMemory.UnSubscribe(sub)
	needUpdateSegment := make(map[int]struct{})
	lastZoneUpdateTime := make(map[zoneNoSegmentTypePair]time.Time)

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

			// for update segment type
			now := time.Now()
			updatedZone := zoneNoSegmentTypePair{ZoneNo: update.ZoneNo, SegmentType: segmentType}
			last := lastZoneUpdateTime[updatedZone]
			// skip send updates if last update of same zoneNo is less than 500ms
			if now.Sub(last) < 500*time.Millisecond {
				continue
			}
			lastZoneUpdateTime[updatedZone] = now
			// notice only zone number
			data = ToZoneResponse(update.ZoneNo, segmentType, nil)
			respBuf.Push(data.Serialize())
		case <-ticker.C:
			if len(needUpdateSegment) == 0 {
				continue
			}
			zone, err = s.znsMemory.GetZone(currentZoneNo)
			if err != nil {
				fmt.Println("Error getting zone", err)
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
			data = ToZoneResponse(currentZoneNo, zone.FrequentSegmentType(), segments)
			respBuf.Push(data.Serialize())
		}
	}
}

func (s *api) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	if !query.Has("dirPath") {
		http.Error(w, "dirPath parameter is required", http.StatusBadRequest)
		return
	}
	dirPath := query.Get("dirPath")
	zoneInfo := s.znsMemory.GetZoneInfo()
	mountInfo, err := fstool.GetMountInfo(zoneInfo.RegularDeviceName)
	if err != nil {
		http.Error(w, "Error getting mount info", http.StatusInternalServerError)
		return
	}

	if dirPath == "" {
		// path 가 빈 경우 root mount path 를 반환
		response := NewListFilesResponse()
		for _, mountPath := range mountInfo.MountPath {
			response.Files = append(response.Files, ListFileItem{
				FilePath: mountPath,
				Name:     mountPath,
				Type:     int(fstool.RootPath),
				SizeStr:  "",
			})
		}
		WriteJsonResponse(w, response)
		return
	}

	// 특정한 path 가 주어진경우
	files, err := fstool.ListFiles(dirPath)
	if err != nil {
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	response := NewListFilesResponse()

	mountPoint := ""
	for _, mountPath := range mountInfo.MountPath {
		if strings.HasPrefix(dirPath, mountPath) {
			mountPoint = mountPath
			break
		}
	}
	if len(mountPoint) > 0 {
		response.MountPoint = mountPoint
		currentDirs := []string{mountPoint}
		remainPath := dirPath[len(mountPoint):]
		if len(remainPath) > 0 && remainPath[0] == '/' {
			remainPath = remainPath[1:]
		}
		if len(remainPath) > 0 {
			currentDirs = append(currentDirs, strings.Split(remainPath, "/")...)
		}
		response.CurrentDirs = currentDirs
		response.ParentDirPath = strings.Join(currentDirs[:len(currentDirs)-1], "/")
	}

	for _, file := range files {
		response.Files = append(response.Files, ListFileItem{
			FilePath: file.FilePath,
			Name:     file.Name,
			Type:     int(file.Type),
			SizeStr:  file.SizeStr,
		})
	}

	WriteJsonResponse(w, response)
	return
}

func (s *api) getFileInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	if !query.Has("filePath") {
		http.Error(w, "filePath parameter is required", http.StatusBadRequest)
		return
	}
	filePath := query.Get("filePath")
	znsInfo := s.znsMemory.GetZoneInfo()
	fileInfo, err := znsmemory.GetFileInfo(znsInfo, filePath)
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	segmentSize := znsInfo.BlockPerSegment / 8 // Assuming 512 bits/8 = 64 bytes
	zoneBlocks := make(map[int][]byte)

	for _, segmentInfo := range fileInfo.FileSegments {
		zoneBlock, ok := zoneBlocks[segmentInfo.ZoneIndex]
		if !ok {
			// Create a new block with the correct size, pre-filled with zeros
			zoneBlock = make([]byte, znsInfo.TotalSegmentPerZone*segmentSize)
			zoneBlocks[segmentInfo.ZoneIndex] = zoneBlock
		}

		if len(segmentInfo.ValidMap) > 0 {
			// Calculate the start position of the segment in the zoneBlock
			i := segmentInfo.RelativeSegmentIndex * segmentSize
			// Copy the data into the zoneBlock
			copy(zoneBlock[i:], segmentInfo.ValidMap)
			// There is no need to reassign the slice to the map because we're modifying the contents directly, not the slice header.
		}
		// If the segment is empty, we leave the pre-filled zeros in place.
	}

	maxHistogramSector := znsInfo.MaxSectors / 4096
	maxHistogramSectorKey := fmt.Sprintf("<%d", maxHistogramSector)
	histogram := make(map[string]int)
	for _, fibmap := range fileInfo.Fibmaps {
		if fibmap.Blks > maxHistogramSector {
			histogram[maxHistogramSectorKey] = histogram[maxHistogramSectorKey] + 1
			continue
		}
		k := strconv.Itoa(fibmap.Blks)
		histogram[k] = histogram[k] + 1
	}

	response := &FileInfoResponse{
		FilePath:       fileInfo.FilePath,
		ZoneBitmaps:    zoneBlocks,
		BlockHistogram: histogram,
	}
	WriteProtoBuf(w, response)
	return
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
	handler.HandleFunc("/api/files", a.listFilesHandler)
	handler.HandleFunc("/api/fileInfo", a.getFileInfoHandler)
	handler.HandleFunc("/static/", a.staticsHandler)
	handler.HandleFunc("/highlight/", a.highlightHandler)
	handler.HandleFunc("/", a.indexHandler)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	installGracefulShutdown(ctx, server)
	return server
}
