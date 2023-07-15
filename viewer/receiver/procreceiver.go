package receiver

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

var _ ZNSReceiver = (*procReceiver)(nil)

type procReceiver struct {
	memory      znsmemory.ZNSMemory
	isReceiving bool
}

func (p procReceiver) StartReceive(ctx context.Context, r *bufio.Reader) {
	if p.isReceiving {
		panic("already receiving")
	}
	p.isReceiving = true
	go func() {
		err := p.readSegmentBits(ctx, r)
		if err != nil {
			fmt.Printf("read segment_bits: %v\n", err)
		}
		p.isReceiving = false
	}()
}

func (p procReceiver) readSegmentBits(ctx context.Context, r *bufio.Reader) error {
	zoneInfo := p.memory.GetZoneInfo()

	// Skip header
	if _, err := r.ReadBytes('\n'); err != nil {
		return fmt.Errorf("read line: %w", err)
	}
	if _, err := r.ReadBytes('\n'); err != nil {
		return fmt.Errorf("read line: %w", err)
	}

	segmentFullNo := 0
	segmentTypeInt := 0
	read := 0
	for ctx.Err() == nil {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read segment bit line: %w", err)
		}
		if len(line) == 0 {
			continue
		}

		read = strings.Index(line, " ")
		if read == -1 {
			return fmt.Errorf("cannot find space in segment bit line")
		}
		intStr := line[:read]
		segmentFullNo, err = strconv.Atoi(intStr)
		if err != nil {
			return fmt.Errorf("parse segment no: %w", err)
		}
		line = line[read+1:]

		read = strings.Index(line, "|")
		if read == -1 {
			return fmt.Errorf("cannot find | in segment bit line")
		}
		segmentTypeStr := line[:read]
		segmentTypeStr = strings.TrimLeft(segmentTypeStr, " ")
		line = line[read+1:]
		segmentTypeInt, err = strconv.Atoi(segmentTypeStr)
		if err != nil {
			return fmt.Errorf("parse segment type: %w", err)
		}
		segmentType := znsmemory.SegmentType(segmentTypeInt)
		if !segmentType.IsValid() {
			return fmt.Errorf("invalid segment type: %d", segmentType)
		}

		// Skip `|`
		read = strings.Index(line, "|")
		if read == -1 {
			return fmt.Errorf("cannot find | in segment bit line")
		}
		line = line[read+2 : len(line)-1] // skip extra space and last newline
		hexStr := strings.Replace(line, " ", "", -1)
		hexBytes, err := hex.DecodeString(hexStr)
		if err != nil {
			return fmt.Errorf("decode hex string: %w", err)
		}
		if len(hexBytes) != validMapSize {
			return fmt.Errorf("invalid valid bitmap size: %d", len(hexBytes))
		}
		zoneNo := segmentFullNo / zoneInfo.TotalSegmentPerZone
		p.memory.UpdateSegment(&znsmemory.SitEntryUpdate{
			SegmentFullNo: segmentFullNo,
			ZoneNo:        zoneNo,
			SegmentType:   segmentType,
			ValidMap:      hexBytes,
		})
	}
	return nil
}

func NewProcReceiver(memory znsmemory.ZNSMemory) ZNSReceiver {
	return &procReceiver{
		memory:      memory,
		isReceiving: false,
	}
}
