package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type SegmentResponse struct {
	Time      int64 `json:"time"`
	ZoneNo    int   `json:"zone_no"`
	SegmentNo int   `json:"segment_no"`
	Map       []int `json:"map"`
}

func (s *SegmentResponse) Serialize() []byte {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(s); err != nil {
		panic(fmt.Errorf("error serializing segment response: %v", err))
	}
	return buf.Bytes()
}
