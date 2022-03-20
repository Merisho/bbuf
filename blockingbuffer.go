package bbuf

import (
	"bytes"
	"errors"
	"sync"
)

var (
	// ErrBufferOverflow indicates that buffer cannot be written to because current buffer bytes + bytes to write > max buffer size
	ErrBufferOverflow = errors.New("buffer overflow")
)

// NewBlockingBuffer constructs BlockingBuffer
func NewBlockingBuffer(size int) *BlockingBuffer {
	bb := &BlockingBuffer{
		buf:     bytes.NewBuffer(make([]byte, 0, size)),
		size:    size,
		bufCond: sync.NewCond(&sync.Mutex{}),
	}
	return bb
}

// BlockingBuffer is a buffer that enables asynchronous write and synchronous read.
// In case buffer is empty, reads will block.
// In case buffer is full, writes will return ErrBufferOverflow
type BlockingBuffer struct {
	buf     *bytes.Buffer
	bufCond *sync.Cond
	size    int
}

// Read reads from the buffer.
// In case buffer is empty, the Read will block
func (bb *BlockingBuffer) Read(b []byte) (int, error) {
	bb.bufCond.L.Lock()
	for bb.buf.Len() == 0 {
		bb.bufCond.Wait()
	}
	n, err := bb.buf.Read(b)
	bb.bufCond.L.Unlock()

	return n, err
}

// Write writes to the buffer.
// In case buffer is full, the Write will return ErrBufferOverflow
func (bb *BlockingBuffer) Write(b []byte) (int, error) {
	bb.bufCond.L.Lock()
	defer bb.bufCond.L.Unlock()
	defer bb.bufCond.Broadcast()

	if bb.buf.Len()+len(b) > bb.size {
		return 0, ErrBufferOverflow
	}

	return bb.buf.Write(b)
}
