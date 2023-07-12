package znsmemory

import (
	"bufio"
	"fmt"
	"io"
)

const validMapSize = SegmentSize / 8

func ReadZoneInfo(r *bufio.Reader) (*ZoneInfo, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read zone info: %w", err)
	}
	var totalZone, zoneBlocks int
	_, err = fmt.Sscanf(line, "info: total_zone=%d zone_blocks=%d", &totalZone, &zoneBlocks)
	if err != nil {
		return nil, fmt.Errorf("parseZoneInfo: %w", err)
	}
	zoneCapBlocks := zoneBlocks // TODO: get real zoneCapBlocks someday
	return &ZoneInfo{
		TotalZone:               totalZone,
		BlockPerSegment:         SegmentSize,
		TotalBlockPerZone:       zoneBlocks,
		AvailableBlockPerZone:   zoneCapBlocks,
		TotalSegmentPerZone:     zoneBlocks / SegmentSize,
		AvailableSegmentPerZone: int(float32(zoneCapBlocks)/float32(SegmentSize) + 0.5),
	}, nil
}

func ReadSitEntryUpdate(r *bufio.Reader) (*SitEntryUpdate, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read update_sit_entry: %w", err)
	}
	var segmentNum, curZone int
	var segmentType SegmentType
	_, err = fmt.Sscanf(line, "update_sit_entry segno: %d cur_zone:%d seg_type:%d", &segmentNum, &curZone, &segmentType)
	if err != nil {
		return nil, fmt.Errorf("parse update_sit_entry: %w", err)
	}
	if !segmentType.IsValid() {
		segmentType = UnknownSegment
	}

	// 64 bytes of validMap and newline
	buf := make([]byte, validMapSize+1) // 64 bytes + 1 newline
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("read update_sit_entry: %w", err)
	}
	return &SitEntryUpdate{
		ZoneNo:      curZone,
		SegmentNo:   segmentNum,
		ValidMap:    buf[:validMapSize],
		SegmentType: segmentType,
	}, nil
}
