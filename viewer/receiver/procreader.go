package receiver

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

func ReadProcSegmentBits(ctx context.Context, memory znsmemory.ZNSMemory, path string) {
	procFile, err := os.Open(path)
	if err != nil {
		fmt.Printf("open %s: %v\n", path, err)
		return
	}
	procReader := bufio.NewReaderSize(procFile, 4096)
	NewProcReceiver(memory).StartReceive(ctx, procReader)
}

func WatchProcSegmentBits(ctx context.Context, sig chan os.Signal, memory znsmemory.ZNSMemory, path string) {
	var lastUpdate, lastReset time.Time
	tick := time.NewTicker(5 * time.Second)
	sub := memory.Subscribe()
	defer memory.UnSubscribe(sub)
	for {
		select {
		case <-ctx.Done():
			return
		case <-sig:
			fmt.Println("reset signal received ...")
			lastReset = time.Now()
			ReadProcSegmentBits(ctx, memory, path)
		case <-sub.Event:
			lastUpdate = time.Now()
		case <-tick.C:
			if !lastUpdate.Equal(lastReset) && time.Since(lastUpdate) > 5*time.Second {
				fmt.Println("no update for 5 seconds, reset ...")
				now := time.Now()
				lastReset = now
				lastUpdate = now
				ReadProcSegmentBits(ctx, memory, path)
			}
		}
	}
}
