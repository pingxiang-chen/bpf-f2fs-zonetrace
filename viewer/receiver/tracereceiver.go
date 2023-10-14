package receiver

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

var _ ZNSReceiver = (*traceReceiver)(nil)

// traceReceiver implements the ZNSReceiver interface and reads segment updates from zone-tracer.
type traceReceiver struct {
	memory      znsmemory.ZNSMemory
	isReceiving bool
}

// StartReceive starts receiving segment updates from the zone-tracer.
func (t *traceReceiver) StartReceive(ctx context.Context, r *bufio.Reader) {
	if t.isReceiving {
		panic("already receiving trace")
	}
	t.isReceiving = true
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			u, err := t.readSitEntryUpdate(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("input closed")
					return
				}
				fmt.Printf("readSegment: %v\n", err)
				continue
			}
			t.memory.UpdateSegment(u)
		}
	}()
}

// readSitEntryUpdate reads and parses a SitEntryUpdate.
func (t *traceReceiver) readSitEntryUpdate(r *bufio.Reader) (*znsmemory.SitEntryUpdate, error) {
	var err error
	intBuf := make([]byte, 4)

	if _, err = io.ReadFull(r, intBuf); err != nil {
		return nil, fmt.Errorf("read segmentNum: %w", err)
	}
	segmentNum := int(binary.LittleEndian.Uint32(intBuf))

	if _, err = io.ReadFull(r, intBuf); err != nil {
		return nil, fmt.Errorf("read curZone: %w", err)
	}
	curZone := int(binary.LittleEndian.Uint32(intBuf))

	if _, err = io.ReadFull(r, intBuf); err != nil {
		return nil, fmt.Errorf("read segmentType: %w", err)
	}
	segmentType := znsmemory.SegmentType(binary.LittleEndian.Uint32(intBuf))
	if !segmentType.IsValid() {
		segmentType = znsmemory.UnknownSegment
	}

	validMap := make([]byte, validMapSize)
	_, err = io.ReadFull(r, validMap)
	if err != nil {
		return nil, fmt.Errorf("read update_sit_entry: %w", err)
	}
	return &znsmemory.SitEntryUpdate{
		ZoneNo:        curZone,
		SegmentFullNo: segmentNum,
		ValidMap:      validMap,
		SegmentType:   segmentType,
	}, nil
}

// ReadZoneInfo reads and parses zone information
func ReadZoneInfo(r *bufio.Reader) (*znsmemory.ZoneInfo, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read zone info: %w", err)
	}
	var totalZone, zoneBlocks int
	mountPath := ""
	_, err = fmt.Sscanf(line, "info: mount=%s total_zone=%d zone_blocks=%d", &mountPath, &totalZone, &zoneBlocks)
	if err != nil {
		return nil, fmt.Errorf("parseZoneInfo: %w (read: %s)", err, line)
	}
	zoneCapBlocks := zoneBlocks // TODO: get real zoneCapBlocks someday
	return &znsmemory.ZoneInfo{
		Device:                  mountPath,
		TotalZone:               totalZone,
		BlockPerSegment:         SegmentSize,
		TotalBlockPerZone:       zoneBlocks,
		AvailableBlockPerZone:   zoneCapBlocks,
		TotalSegmentPerZone:     zoneBlocks / SegmentSize,
		AvailableSegmentPerZone: int(float32(zoneCapBlocks)/float32(SegmentSize) + 0.5),
	}, nil
}

// NewTraceReceiver creates a new instance of ZNSReceiver for reading trace updates.
func NewTraceReceiver(memory znsmemory.ZNSMemory) ZNSReceiver {
	return &traceReceiver{
		memory:      memory,
		isReceiving: false,
	}
}
