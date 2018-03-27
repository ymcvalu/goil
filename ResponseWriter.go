package goil

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.CloseNotifier
	http.Hijacker
	http.Flusher
}

var _ ResponseWriter = new(responseWriter)

type responseWriter struct {
	writer http.ResponseWriter
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (conn net.Conn, io *bufio.ReadWriter, err error) {
	if hijacker, ok := w.writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	err = errors.New("the recevier isn't a Hijacker")
	return
}

func (w *responseWriter) CloseNotify() <-chan bool {
	if closeNotifier, ok := w.writer.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}

func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *responseWriter) Write(bytes []byte) (int, error) {
	return w.Write(bytes)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.WriteHeader(statusCode)
}
