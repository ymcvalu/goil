package goil

import (
	"bufio"
	"net"
	"net/http"
)

const (
	nowriten int64 = -1
)

type Response interface {
	http.ResponseWriter
	http.CloseNotifier
	http.Hijacker
	http.Flusher
	Status() int
	Size() int64
	reset(writer http.ResponseWriter)
	clear() interface{}
	SetHeader(key, value string)
}

//assert that response implements Response
var _ Response = new(response)

type response struct {
	writer http.ResponseWriter
	status int
	size   int64
}

//http.ResponseWriter implements Flusher
//panic if type assert failed
//force flush to conn
func (w *response) Flush() {
	w.writer.(http.Flusher).Flush()
}

//http.ResponseWriter implements Hijacker
//panic if type assert failed
//get the connection under request
func (w *response) Hijack() (conn net.Conn, io *bufio.ReadWriter, err error) {
	return w.writer.(http.Hijacker).Hijack()
}

//http.ResponseWriter implements CloseNotifier
//panic if type assert failed
//get a notifier for connection closing
func (w *response) CloseNotify() <-chan bool {
	return w.writer.(http.CloseNotifier).CloseNotify()
}

func (w *response) Header() http.Header {
	return w.writer.Header()
}

func (w *response) SetHeader(key, value string) {
	w.Header().Add(key, value)
}

//if the btyes is nil or length is zero, and no body wrote,then the func will send
// a response header to the client
func (w *response) Write(bytes []byte) (n int, err error) {
	if w.size == nowriten {
		w.writer.WriteHeader(w.status)
		w.size = 0
	}
	if bytes != nil && len(bytes) > 0 && bodyAllowedForStatus(w.status) {
		n, err = w.writer.Write(bytes)
		w.size += int64(n)
	}
	return
}

//WriteHeader set the response status code
func (w *response) WriteHeader(statusCode int) {
	if w.size != nowriten {
		//TODO:add log
		w.status = statusCode
	}
}

//get the current status
func (w *response) Status() int {
	return w.status
}

//get the size that has wrote to connection
func (w *response) Size() int64 {
	return w.size
}

//reset a response for reuse
//the status default is 200
func (w *response) reset(writer http.ResponseWriter) {
	w.writer = writer
	w.size = nowriten
	//TODO:replace http code from http to goil
	w.status = http.StatusOK
}

func (w *response) clear() interface{} {
	w.writer = nil
	return w
}

func newResponse() Response {
	return &response{
		status: http.StatusOK,
	}
}

//copy from net/http package, unexport method
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
