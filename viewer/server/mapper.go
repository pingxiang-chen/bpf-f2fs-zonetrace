package server

import (
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
	"time"
)

func ToZoneInfoResponse(info znsmemory.ZoneInfo) ZoneInfoResponse {
	return ZoneInfoResponse{
		TotalZone:               info.TotalZone,
		BlockPerSegment:         info.BlockPerSegment,
		TotalBlockPerZone:       info.TotalBlockPerZone,
		AvailableBlockPerZone:   info.AvailableBlockPerZone,
		TotalSegmentPerZone:     info.TotalSegmentPerZone,
		AvailableSegmentPerZone: info.AvailableSegmentPerZone,
	}
}

func ToSegmentResponse(zoneNo int, segmentNo int, validMap znsmemory.ValidMap) SegmentResponse {
	t := time.Now().UnixNano() / int64(time.Millisecond) // unix time in ms
	if len(validMap) == 0 {
		return SegmentResponse{
			Time:      t,
			ZoneNo:    zoneNo,
			SegmentNo: segmentNo,
			Map:       nil,
		}
	}
	row := make([]int, 512)
	rowIndex := 0
	for _, b := range validMap {
		for i := 0; i < 8; i++ {
			if b&(1<<uint(i)) != 0 {
				row[rowIndex] = 1
			} else {
				row[rowIndex] = 0
			}
			rowIndex++
		}
	}
	return SegmentResponse{
		Time:      t,
		ZoneNo:    zoneNo,
		SegmentNo: segmentNo,
		Map:       row,
	}
}
