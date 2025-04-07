package response_writer

import (
	"bytes"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		statusCode:     http.StatusOK,
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *ResponseWriter) Reset() {
	rw.body.Reset()
	rw.statusCode = http.StatusOK
}

func (rw *ResponseWriter) Flush() error {
	rw.ResponseWriter.WriteHeader(rw.statusCode)
	_, err := rw.ResponseWriter.Write(rw.body.Bytes())
	return err
}
