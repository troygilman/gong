package gong

import (
	"bytes"
	"net/http"
)

type CustomResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		statusCode:     http.StatusOK,
	}
}

func (crw *CustomResponseWriter) Header() http.Header {
	return crw.ResponseWriter.Header()
}

func (crw *CustomResponseWriter) Write(b []byte) (int, error) {
	return crw.body.Write(b)
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
}

func (crw *CustomResponseWriter) Reset() {
	crw.body.Reset()
	crw.statusCode = http.StatusOK
}

func (crw *CustomResponseWriter) Flush() error {
	crw.ResponseWriter.WriteHeader(crw.statusCode)
	_, err := crw.ResponseWriter.Write(crw.body.Bytes())
	return err
}
