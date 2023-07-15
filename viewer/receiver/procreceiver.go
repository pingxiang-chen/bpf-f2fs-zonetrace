package receiver

import (
	"bufio"
	"context"
)

var _ ZNSReceiver = (*procReceiver)(nil)

type procReceiver struct {
	isReceiving bool
}

func (p procReceiver) StartReceive(ctx context.Context, r *bufio.Reader) {
	if p.isReceiving {
		panic("already receiving")
	}
	panic("implement me")
}
