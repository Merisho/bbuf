# Bbuf — the blocking buffer
Bbuf wraps `bytes.Buffer` to provide asynchronous writes and synchronous reads.

If the buffer is empty - all reads block. If the buffer is full - all writes return an error.

Safe for concurrent use.
