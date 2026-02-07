package proxy

import (
	"io"
	"net/http"
)

// StreamResponse streams an HTTP response body to a writer
func StreamResponse(w io.Writer, resp *http.Response) error {
	defer resp.Body.Close()
	_, err := io.Copy(w, resp.Body)
	return err
}

// StreamRequest streams an HTTP request body to a writer
func StreamRequest(w io.Writer, req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	defer req.Body.Close()
	_, err := io.Copy(w, req.Body)
	return err
}

// CopyHeaders copies HTTP headers from source to destination
func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// StreamingBufferSize defines the buffer size for streaming operations
const StreamingBufferSize = 32 * 1024 // 32KB

// BufferedCopy copies data from reader to writer with a specific buffer size
func BufferedCopy(dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, StreamingBufferSize)
	return io.CopyBuffer(dst, src, buf)
}
