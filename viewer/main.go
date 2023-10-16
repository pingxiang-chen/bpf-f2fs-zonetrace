package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/receiver"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/server"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

// newSignalContext creates a new context with a cancel function that's triggered
// when an interrupt signal (e.g., Ctrl+C) is received.
func newSignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		cancel()
	}()
	return ctx
}

// newResetSignal creates a new channel that receives a signal when a SIGHUP signal is received.
func newResetSignal() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	return c
}

func main() {
	// Create a buffered reader for standard input with a buffer size of 4096 bytes.
	stdioReader := bufio.NewReaderSize(os.Stdin, 4096)

	// Read zone information from standard input
	zoneInfo, err := receiver.ReadZoneInfo(stdioReader)
	if err != nil {
		panic(fmt.Errorf("readZoneInfo: %w", err))
	}

	// Open the procFile for reading to read F2FS segment bits information.
	procPath := fmt.Sprintf("/proc/fs/f2fs/%s/segment_bits", zoneInfo.RegularDeviceName)
	isProcFileExist := false
	if _, err := os.Stat(procPath); err == nil {
		isProcFileExist = true
	}

	ctx := newSignalContext()          // Create a signal context for managing signals.
	m := znsmemory.New(ctx, *zoneInfo) // Create a new in-memory store to save all updates.

	// If procFile exists, start receiving segment bits from it.
	if isProcFileExist {
		resetSignal := newResetSignal()
		receiver.ReadProcSegmentBits(ctx, m, procPath)
		go receiver.WatchProcSegmentBits(ctx, resetSignal, m, procPath)
	}

	port := 9090
	// Start receiving traces from standard input.
	receiver.NewTraceReceiver(m).StartReceive(ctx, stdioReader)

	// Create an HTTP server
	srv := server.New(ctx, m, port)
	fmt.Printf("======== Running on http://0.0.0.0:%d ========\n", port)
	// Start the HTTP server and wait for it to close.
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
