package receiver

import (
	"bufio"
	"context"
)

type ZNSReceiver interface {
	StartReceive(ctx context.Context, r *bufio.Reader)
}
