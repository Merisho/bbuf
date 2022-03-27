# Bbuf — the blocking buffer
Bbuf wraps `Buffer` interface to provide asynchronous writes and synchronous reads.
`Buffer` interface is implemented by `bytes.Buffer` out of the box.

If the buffer is empty - all reads block. If the buffer is full - all writes return an error.

In case the underlying `Buffer` has 0 capacity, all reads block and all writes return an error.

Safe for concurrent use.

## Example
```
b := bytes.NewBuffer(make([]byte, 0, 4 * 1024))
bb := bbuf.New(b)

go func() {
    time.Sleep(5 * time.Second)
    _, err := bb.Write(bytes.Repeat([]byte{1}, 4 * 1024))
    if err != nil {
        panic(err)
    }
}()

data := make([]byte, 4 * 1024)
fmt.Println("reading...")
n, err := bb.Read(data) // blocks until buffer has data
fmt.Println("read", n, err, data[0], data[4 * 1024 - 1]) // read 4096 <nil> 1 1
```
