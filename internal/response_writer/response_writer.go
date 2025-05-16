package response_writer

import (
	"bytes"
	"net/http"
)

// ResponseWriter is a buffered HTTP response writer that captures response
// content before writing it to the underlying http.ResponseWriter.
// This allows for deferred writes and response manipulation before sending.
// It's used by Gong to handle HTMX requests and enable features like
// out-of-band swaps and deferred rendering.
type ResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter that wraps the provided
// http.ResponseWriter with buffering capabilities.
// The buffer is initially empty and the status code is set to 200 OK.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		statusCode:     http.StatusOK,
	}
}

// Header returns the header map that will be sent to the client.
// This is a passthrough to the underlying http.ResponseWriter's Header method.
func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// Write adds the provided bytes to the response buffer.
// The response is not sent until Flush is called.
// This allows Gong to accumulate response data before deciding how to send it.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

// WriteHeader sets the status code for the response.
// The status code is not sent until Flush is called.
// This allows Gong to set a status code that will be used when the response is flushed.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Reset clears the response buffer and resets the status code to 200 OK.
// This is useful when an action needs to discard its current response and start over.
func (rw *ResponseWriter) Reset() {
	rw.body.Reset()
	rw.statusCode = http.StatusOK
}

// Flush writes the buffered response to the underlying ResponseWriter.
// This sends the status code, headers, and body to the client.
// It's called automatically at the end of request processing by the Gong server.
func (rw *ResponseWriter) Flush() error {
	rw.ResponseWriter.WriteHeader(rw.statusCode)
	_, err := rw.ResponseWriter.Write(rw.body.Bytes())
	return err
}
