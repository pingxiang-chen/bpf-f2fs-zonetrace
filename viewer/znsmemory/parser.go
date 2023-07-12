package znsmemory

import (
	"bufio"
	"encoding/binary"
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
	var err error
	intBuf := make([]byte, 4)

	if _, err = r.Read(intBuf); err != nil {
		return nil, fmt.Errorf("read segmentNum: %w", err)
	}
	segmentNum := int(binary.LittleEndian.Uint32(intBuf))

	if _, err = r.Read(intBuf); err != nil {
		return nil, fmt.Errorf("read curZone: %w", err)
	}
	curZone := int(binary.LittleEndian.Uint32(intBuf))

	if _, err = r.Read(intBuf); err != nil {
		return nil, fmt.Errorf("read segmentType: %w", err)
	}
	segmentType := SegmentType(binary.LittleEndian.Uint32(intBuf))
	if !segmentType.IsValid() {
		segmentType = UnknownSegment
	}

	validMap := make([]byte, validMapSize)
	_, err = io.ReadFull(r, validMap)
	if err != nil {
		return nil, fmt.Errorf("read update_sit_entry: %w", err)
	}
	return &SitEntryUpdate{
		ZoneNo:      curZone,
		SegmentNo:   segmentNum,
		ValidMap:    validMap,
		SegmentType: segmentType,
	}, nil
}
