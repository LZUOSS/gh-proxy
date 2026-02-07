package util

import (
	"io"
	"sync"
)

// CopyBuffer copies from src to dst using a provided buffer.
// This is a wrapper around io.CopyBuffer that provides additional functionality.
func CopyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024) // 32KB default buffer
	}
	return io.CopyBuffer(dst, src, buf)
}

// MultiWriter creates a writer that duplicates its writes to all provided writers.
// This is useful for writing to multiple destinations simultaneously.
func MultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// LimitedReader wraps an io.Reader and limits the number of bytes that can be read.
type LimitedReader struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

// Read reads up to len(p) bytes into p
func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

// NewLimitedReader returns a Reader that reads from r but stops with EOF after n bytes.
func NewLimitedReader(r io.Reader, n int64) *LimitedReader {
	return &LimitedReader{R: r, N: n}
}

// BufferPool is a pool of byte buffers for efficient memory reuse
type BufferPool struct {
	pool sync.Pool
	size int
}

// NewBufferPool creates a new buffer pool with the given buffer size
func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, size)
				return &buf
			},
		},
		size: size,
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() *[]byte {
	return p.pool.Get().(*[]byte)
}

// Put returns a buffer to the pool
func (p *BufferPool) Put(buf *[]byte) {
	if buf == nil {
		return
	}
	p.pool.Put(buf)
}

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with corresponding writes to w.
func TeeReader(r io.Reader, w io.Writer) io.Reader {
	return io.TeeReader(r, w)
}

// CountingWriter wraps an io.Writer and counts the bytes written
type CountingWriter struct {
	Writer       io.Writer
	BytesWritten int64
}

// Write writes p to the underlying writer and counts bytes
func (cw *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	cw.BytesWritten += int64(n)
	return
}

// NewCountingWriter creates a new CountingWriter
func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{Writer: w}
}

// CountingReader wraps an io.Reader and counts the bytes read
type CountingReader struct {
	Reader   io.Reader
	BytesRead int64
}

// Read reads from the underlying reader and counts bytes
func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.BytesRead += int64(n)
	return
}

// NewCountingReader creates a new CountingReader
func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{Reader: r}
}
