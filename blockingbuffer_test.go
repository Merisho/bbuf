package bbuf

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestBlockingBuffer(t *testing.T) {
	suite.Run(t, new(BlockingBufferTestSuite))
}

type BlockingBufferTestSuite struct {
	suite.Suite
}

func (ts *BlockingBufferTestSuite) TestNonBlockingWrite() {
	bb := NewBlockingBuffer(1024)

	n, err := bb.Write(make([]byte, 128))
	ts.NoError(err)
	ts.Equal(128, n)
}

func (ts *BlockingBufferTestSuite) TestBlockingRead() {
	bb := NewBlockingBuffer(1024)

	_, err := ts.receiveWithTimeout(ts.asyncRead(bb, 64))
	ts.EqualError(err, "timeout", "timeout must occur because read must block when buffer is empty")
}

func (ts *BlockingBufferTestSuite) TestWriteRead() {
	bb := NewBlockingBuffer(1024)

	_, err := bb.Write(make([]byte, 128))
	ts.NoError(err)

	n, err := bb.Read(make([]byte, 64))
	ts.NoError(err)
	ts.Equal(64, n)

	n, err = bb.Read(make([]byte, 64))
	ts.NoError(err)
	ts.Equal(64, n)

	_, err = ts.receiveWithTimeout(ts.asyncRead(bb, 64))
	ts.EqualError(err, "timeout", "timeout must occur because read must block when buffer is empty")
}

func (ts *BlockingBufferTestSuite) TestWriteWhenReadIsBlocked() {
	bb := NewBlockingBuffer(1024)

	readRes := ts.asyncRead(bb, 64)

	n, err := bb.Write(make([]byte, 1024))
	ts.NoError(err)
	ts.Equal(1024, n)

	n, err = ts.receiveWithTimeout(readRes)
	ts.NoError(err, "write must unblock waiting read")
	ts.Equal(64, n)
}

func (ts *BlockingBufferTestSuite) TestConcurrentReadsUnblockedByConcurrentWrites() {
	bb := NewBlockingBuffer(1024)

	readRes1 := ts.asyncRead(bb, 64)
	readRes2 := ts.asyncRead(bb, 64)

	go func() {
		n, err := bb.Write(make([]byte, 64))
		ts.NoError(err)
		ts.Equal(64, n)
	}()

	go func() {
		n, err := bb.Write(make([]byte, 64))
		ts.NoError(err)
		ts.Equal(64, n)
	}()

	n, err := ts.receiveWithTimeout(readRes1)
	ts.NoError(err)
	ts.Equal(64, n)

	n, err = ts.receiveWithTimeout(readRes2)
	ts.NoError(err)
	ts.Equal(64, n)
}

func (ts *BlockingBufferTestSuite) TestBufferOverflow() {
	bb := NewBlockingBuffer(16)

	n, err := bb.Write(make([]byte, 16))
	ts.NoError(err)
	ts.Equal(16, n)

	n, err = bb.Write(make([]byte, 16))
	ts.Error(err, "buffer overflow")
	ts.Equal(0, n)
}

type readResult struct {
	n   int
	err error
}

func (ts *BlockingBufferTestSuite) asyncRead(bb *BlockingBuffer, size int) chan readResult {
	res := make(chan readResult)

	go func() {
		n, err := bb.Read(make([]byte, size))
		res <- readResult{
			n:   n,
			err: err,
		}
	}()

	return res
}

func (ts *BlockingBufferTestSuite) receiveWithTimeout(res chan readResult) (n int, err error) {
	select {
	case r := <-res:
		return r.n, r.err
	case <-time.After(10 * time.Millisecond):
		return 0, errors.New("timeout")
	}
}
