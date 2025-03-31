package gong

import (
	"bytes"
	"net/http"
)

type CustomResponseWriter struct {
	http.ResponseWriter
	body          *bytes.Buffer
	statusCode    int
	headerWritten bool
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
	if !crw.headerWritten {
		crw.WriteHeader(crw.statusCode)
	}
	// crw.body.Write(b)
	return crw.ResponseWriter.Write(b)
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	if !crw.headerWritten {
		crw.statusCode = statusCode
		crw.ResponseWriter.WriteHeader(statusCode)
		crw.headerWritten = true
	}
}

func (crw *CustomResponseWriter) Reset() {
	// crw.body.Reset()
	crw.statusCode = http.StatusOK
	crw.headerWritten = false
}
