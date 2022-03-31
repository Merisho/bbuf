package bbuf

import (
	"errors"
	"io"
	"sync"
)

var (
	// ErrBufferOverflow indicates that buffer cannot be written to because current buffer bytes + bytes to write > max buffer size
	ErrBufferOverflow = errors.New("buffer overflow")

	ErrBufferClosed = errors.New("buffer closed")
)

type Buffer interface {
	io.ReadWriter

	Len() int
	Cap() int
}

// New constructs BlockingBuffer
func New(buf Buffer) *BlockingBuffer {
	bb := &BlockingBuffer{
		buf:     buf,
		bufCond: sync.NewCond(&sync.Mutex{}),
	}
	return bb
}

// BlockingBuffer is a buffer that enables asynchronous write and synchronous read.
// In case buffer is empty, reads will block.
// In case buffer is full, writes will return ErrBufferOverflow
type BlockingBuffer struct {
	buf     Buffer
	bufCond *sync.Cond
	size    int
}

// Read reads from the buffer.
// In case buffer is empty, the Read will block
func (bb *BlockingBuffer) Read(b []byte) (int, error) {
	bb.bufCond.L.Lock()
	for bb.buf != nil && bb.buf.Len() == 0 {
		bb.bufCond.Wait()
	}

	defer bb.bufCond.L.Unlock()

	if bb.buf == nil {
		return 0, io.EOF
	}

	n, err := bb.buf.Read(b)

	return n, err
}

// Write writes to the buffer.
// In case buffer is full, Write will return ErrBufferOverflow
func (bb *BlockingBuffer) Write(b []byte) (int, error) {
	bb.bufCond.L.Lock()
	defer bb.bufCond.L.Unlock()

	if bb.buf == nil {
		return 0, ErrBufferClosed
	}

	defer bb.bufCond.Broadcast()

	if bb.buf.Len()+len(b) > bb.buf.Cap() {
		return 0, ErrBufferOverflow
	}

	return bb.buf.Write(b)
}

// Close closes the buffer and makes it unusable
func (bb *BlockingBuffer) Close() error {
	bb.bufCond.L.Lock()
	defer bb.bufCond.L.Unlock()
	defer bb.bufCond.Broadcast()

	bb.buf = nil

	return nil
}
