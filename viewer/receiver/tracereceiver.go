package receiver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

var _ ZNSReceiver = (*traceReceiver)(nil)

type traceReceiver struct {
	memory      znsmemory.ZNSMemory
	isReceiving bool
}

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
			u, err := ReadSitEntryUpdate(r)
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

func NewTraceReceiver(memory znsmemory.ZNSMemory) ZNSReceiver {
	return &traceReceiver{
		memory:      memory,
		isReceiving: false,
	}
}
