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

func main() {
	stdioReader := bufio.NewReaderSize(os.Stdin, 4096)
	zoneInfo, err := receiver.ReadZoneInfo(stdioReader)
	if err != nil {
		panic(fmt.Errorf("readZoneInfo: %w", err))
	}

	procPath := fmt.Sprintf("/proc/fs/f2fs/%s/segment_bits", zoneInfo.DeviceName)
	procFile, err := os.Open(procPath)
	if err != nil {
		fmt.Printf("open %s: %v\n", procPath, err)
	}

	port := 9090
	ctx := newSignalContext()
	m := znsmemory.New(ctx, *zoneInfo)
	receiver.NewTraceReceiver(m).StartReceive(ctx, stdioReader)
	if procFile != nil {
		procReader := bufio.NewReaderSize(procFile, 4096)
		receiver.NewProcReceiver(m).StartReceive(ctx, procReader)
	}
	srv := server.New(ctx, m, port)
	fmt.Printf("======== Running on http://0.0.0.0:%d ========\n", port)
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
