package bbuf

import (
	"bytes"
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
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	bb := New(b)

	n, err := bb.Write(make([]byte, 128))
	ts.NoError(err)
	ts.Equal(128, n)
}

func (ts *BlockingBufferTestSuite) TestBlockingRead() {
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	bb := New(b)

	_, err := ts.receiveWithTimeout(ts.asyncRead(bb, 64))
	ts.EqualError(err, "timeout", "timeout must occur because read must block when buffer is empty")
}

func (ts *BlockingBufferTestSuite) TestWriteRead() {
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	bb := New(b)

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
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	bb := New(b)

	readRes := ts.asyncRead(bb, 64)

	n, err := bb.Write(make([]byte, 1024))
	ts.NoError(err)
	ts.Equal(1024, n)

	n, err = ts.receiveWithTimeout(readRes)
	ts.NoError(err, "write must unblock waiting read")
	ts.Equal(64, n)
}

func (ts *BlockingBufferTestSuite) TestConcurrentReadsUnblockedByConcurrentWrites() {
	b := bytes.NewBuffer(make([]byte, 0, 1024))
	bb := New(b)

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
	b := bytes.NewBuffer(make([]byte, 0, 16))
	bb := New(b)

	n, err := bb.Write(make([]byte, 16))
	ts.NoError(err)
	ts.Equal(16, n)

	n, err = bb.Write(make([]byte, 16))
	ts.Error(err, "buffer overflow")
	ts.Equal(0, n)
}

func (ts *BlockingBufferTestSuite) TestBytesBufferWithNilBytesSlice() {
	b := bytes.NewBuffer(nil)
	bb := New(b)

	n, err := bb.Write([]byte{10: 0})
	ts.Zero(n)
	ts.Error(err, "buffer overflow")

	n, err = ts.receiveWithTimeout(ts.asyncRead(bb, 1))
	ts.Zero(n)
	ts.Error(err, "timeout")
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
