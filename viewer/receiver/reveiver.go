package receiver

import (
	"bufio"
	"context"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

const SegmentSize = znsmemory.SegmentSize
const validMapSize = SegmentSize / 8

// ZNSReceiver is an interface for receiving segment information.
// By implementing this, you can receive zns status updates from various sources.
type ZNSReceiver interface {
	StartReceive(ctx context.Context, r *bufio.Reader)
}
