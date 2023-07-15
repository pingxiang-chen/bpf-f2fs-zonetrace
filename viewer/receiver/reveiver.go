package receiver

import (
	"bufio"
	"context"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

const SegmentSize = znsmemory.SegmentSize
const validMapSize = SegmentSize / 8

type ZNSReceiver interface {
	StartReceive(ctx context.Context, r *bufio.Reader)
}
