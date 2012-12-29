package mosquitto

/*
#cgo LDFLAGS: -lmosquitto
#include "mosquitto_ext.h"
*/
import "C"

import (
	"unsafe"
)

type HandlerFunc func(*Conn, Message)

type Handler struct {
	Sub     string
	Qos     int
	handler HandlerFunc
}

func NewHandler(sub string, qos int, handler HandlerFunc) (Handler, error) {
	return Handler{sub, qos, handler}, nil
}

func (h Handler) Call(conn *Conn, message Message) error {
	h.handler(conn, message)
	return nil
}

func (h Handler) Match(topic string) bool {
	ctopic := C.CString(topic)
	csub := C.CString(h.Sub)
	cresult := C.bool(false)
	defer C.free(unsafe.Pointer(ctopic))
	defer C.free(unsafe.Pointer(csub))

	C.mosquitto_topic_matches_sub(csub, ctopic, &cresult)
	if cresult {
		return true
	}
	return false
}
