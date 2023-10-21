package server

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/proto"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/protos"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/rle"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

type ZoneInfoResponse struct {
	TotalZone               int `json:"total_zone"`
	BlockPerSegment         int `json:"block_per_segment"`
	TotalBlockPerZone       int `json:"total_block_per_zone"`
	AvailableBlockPerZone   int `json:"available_block_per_zone"`
	TotalSegmentPerZone     int `json:"total_segment_per_zone"`
	AvailableSegmentPerZone int `json:"available_segment_per_zone"`
	LastSegmentType         int `json:"last_segment_type"`
}

func (z *ZoneInfoResponse) Serialize() []byte {
	b, err := json.Marshal(z)
	if err != nil {
		panic(fmt.Errorf("error serializing zone info response: %v", err))
	}
	return b
}

type Segment struct {
	SegmentNo   int    `json:"segment_no"`
	SegmentType int    `json:"segment_type"`
	Map         []byte `json:"map"`
}

type ZoneResponse struct {
	Time            int64
	ZoneNo          int
	LastSegmentType int
	Segments        []Segment
}

func (z *ZoneResponse) Serialize() []byte {
	p := &protos.ZoneResponse{
		Time:            z.Time,
		ZoneNo:          int32(z.ZoneNo),
		LastSegmentType: int32(z.LastSegmentType),
		Segments:        make([]*protos.Segment, len(z.Segments)),
	}
	for i, segment := range z.Segments {
		p.Segments[i] = &protos.Segment{
			SegmentNo:   int32(segment.SegmentNo),
			Map:         segment.Map,
			SegmentType: int32(segment.SegmentType),
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	_, err := protodelim.MarshalTo(buf, p)
	if err != nil {
		panic(fmt.Errorf("error serializing segment response: %v", err))
	}
	return buf.Bytes()
}

type zoneNoSegmentTypePair struct {
	ZoneNo      int
	SegmentType znsmemory.SegmentType
}

type ListFilesResponse struct {
	MountPoint    string         `json:"mount_point"`
	CurrentDirs   []string       `json:"current_dirs"`
	ParentDirPath string         `json:"parent_dir_path"`
	Files         []ListFileItem `json:"files"`
}

func NewListFilesResponse() *ListFilesResponse {
	return &ListFilesResponse{
		CurrentDirs: make([]string, 0),
		Files:       make([]ListFileItem, 0),
	}
}

type ListFileItem struct {
	FilePath string `json:"file_path"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	SizeStr  string `json:"size_str"`
}

func (r *ListFilesResponse) Serialize() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Printf("error serializing list files response: %v\n", err)
	}
	return b
}

type FileInfoResponse struct {
	FilePath       string         `json:"file_path"`
	ZoneBitmaps    map[int][]byte `json:"zone_bitmaps"`
	BlockHistogram map[string]int `json:"block_histogram"`
}

func (r *FileInfoResponse) Serialize() []byte {
	zoneBitmaps := make(map[int32][]byte)
	for k, v := range r.ZoneBitmaps {
		zoneBitmaps[int32(k)] = rle.Compress(v)
	}
	blockHistogram := make(map[string]int32)
	for k, v := range r.BlockHistogram {
		blockHistogram[k] = int32(v)
	}
	msg := protos.FileInfoResponse{
		FilePath:       r.FilePath,
		ZoneBitmaps:    zoneBitmaps,
		BlockHistogram: blockHistogram,
	}
	b, err := proto.Marshal(&msg)
	if err != nil {
		fmt.Printf("error serializing file info response: %v\n", err)
	}
	return b
}
