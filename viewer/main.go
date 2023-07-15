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
	r := bufio.NewReaderSize(os.Stdin, 4096)
	zoneInfo, err := receiver.ReadZoneInfo(r)
	if err != nil {
		panic(fmt.Errorf("readZoneInfo: %w", err))
	}

	port := 9090
	ctx := newSignalContext()
	m := znsmemory.New(ctx, *zoneInfo)
	receiver.NewTraceReceiver(m).StartReceive(ctx, r)
	srv := server.New(ctx, m, port)
	fmt.Printf("======== Running on http://0.0.0.0:%d ========\n", port)
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
