package server

import (
	"time"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

func ToZoneInfoResponse(info znsmemory.ZoneInfo, lastSegmentType znsmemory.SegmentType) ZoneInfoResponse {
	return ZoneInfoResponse{
		TotalZone:               info.TotalZone,
		BlockPerSegment:         info.BlockPerSegment,
		TotalBlockPerZone:       info.TotalBlockPerZone,
		AvailableBlockPerZone:   info.AvailableBlockPerZone,
		TotalSegmentPerZone:     info.TotalSegmentPerZone,
		AvailableSegmentPerZone: info.AvailableSegmentPerZone,
		LastSegmentType:         int(lastSegmentType),
	}
}

func ToZoneResponse(zoneNo int, segmentType znsmemory.SegmentType, segments []Segment) ZoneResponse {
	t := time.Now().UnixNano() / int64(time.Millisecond) // unix time in ms
	return ZoneResponse{
		Time:            t,
		ZoneNo:          zoneNo,
		LastSegmentType: int(segmentType),
		Segments:        segments,
	}
}
