package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/protos"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
	"google.golang.org/protobuf/encoding/protodelim"
)

type ZoneInfoResponse struct {
	TotalZone               int `json:"total_zone"`
	BlockPerSegment         int `json:"block_per_segment"`
	TotalBlockPerZone       int `json:"total_block_per_zone"`
	AvailableBlockPerZone   int `json:"available_block_per_zone"`
	TotalSegmentPerZone     int `json:"total_segment_per_zone"`
	AvailableSegmentPerZone int `json:"available_segment_per_zone"`
}

func (z *ZoneInfoResponse) Serialize() []byte {
	b, err := json.Marshal(z)
	if err != nil {
		panic(fmt.Errorf("error serializing zone info response: %v", err))
	}
	return b
}

type Segment struct {
	SegmentNo int    `json:"segment_no"`
	Map       []byte `json:"map"`
}

type ZoneResponse struct {
	Time        int64
	ZoneNo      int
	SegmentType int
	Segments    []Segment
}

func (z *ZoneResponse) Serialize() []byte {
	p := &protos.ZoneResponse{
		Time:        z.Time,
		ZoneNo:      int32(z.ZoneNo),
		SegmentType: int32(z.SegmentType),
		Segments:    make([]*protos.Segment, len(z.Segments)),
	}
	for i, segment := range z.Segments {
		p.Segments[i] = &protos.Segment{
			SegmentNo: int32(segment.SegmentNo),
			Map:       segment.Map,
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
