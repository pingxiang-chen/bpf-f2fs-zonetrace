package respbuffer

import "context"

var _ ResponseBuffer = (*responseBuffer)(nil)

// responseBufferSize defines the maximum buffer size.
const responseBufferSize = 128

// ResponseBuffer is an interface that allows you to manage a buffer of responses.
// It ensures that responses do not accumulate beyond a specified responseBufferSize.
// When you push data into the buffer, and it is already full, the oldest response is removed,
// and the new response is added at the end.
type ResponseBuffer interface {
	// Push adds the provided data to the buffer. If the buffer is full, it removes the oldest response.
	Push(data []byte)

	// PopFirst retrieves and returns the oldest(first) response from the buffer.
	PopFirst() ([]byte, error)
}

// responseBuffer is the implementation of the ResponseBuffer interface.
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

// New creates a new instance of ResponseBuffer
func New(ctx context.Context) ResponseBuffer {
	return &responseBuffer{
		ctx:  ctx,
		size: responseBufferSize,
		ch:   make(chan []byte, responseBufferSize),
	}
}
