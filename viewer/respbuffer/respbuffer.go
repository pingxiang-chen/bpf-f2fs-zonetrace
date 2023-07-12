package respbuffer

import "context"

const responseBufferSize = 128

type ResponseBuffer interface {
	Push(data []byte)
	PopFirst() ([]byte, error)
}

type responseBuffer struct {
	ctx  context.Context
	size int
	ch   chan []byte
}

func (b *responseBuffer) Push(data []byte) {
	if len(b.ch) == b.size {
		select {
		case <-b.ch:
		default:
		}
	}
	select {
	case <-b.ctx.Done():
		return
	case b.ch <- data:
		return
	}
}

func (b *responseBuffer) PopFirst() ([]byte, error) {
	select {
	case <-b.ctx.Done():
		return nil, b.ctx.Err()
	case data := <-b.ch:
		return data, nil
	}
}

func New(ctx context.Context) ResponseBuffer {
	return &responseBuffer{
		ctx:  ctx,
		size: responseBufferSize,
		ch:   make(chan []byte, responseBufferSize),
	}
}
