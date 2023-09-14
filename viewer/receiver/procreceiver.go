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

// procReceiver reads segment information from /proc/fs/f2fs/${device}/segment_bits only once.
// It implements the ZNSReceiver interface.
type procReceiver struct {
	memory      znsmemory.ZNSMemory
	isReceiving bool
}

// StartReceive starts receiving current segment information.
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

// readSegmentBits reads and update segment information
func (p procReceiver) readSegmentBits(ctx context.Context, r *bufio.Reader) error {
	zoneInfo := p.memory.GetZoneInfo()

	// Skip header lines
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

		// Parse segment number
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

		// Parse segment type
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

		read = strings.Index(line, "|")
		if read == -1 {
			return fmt.Errorf("cannot find | in segment bit line")
		}
		line = line[read+2 : len(line)-1] // skip `|`, skip extra space and last newline

		// Parse valid map
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

// NewProcReceiver creates a new instance of ZNSReceiver for reading from the /proc filesystem.
func NewProcReceiver(memory znsmemory.ZNSMemory) ZNSReceiver {
	return &procReceiver{
		memory:      memory,
		isReceiving: false,
	}
}
