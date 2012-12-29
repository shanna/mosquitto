package mosquitto

/*
#cgo LDFLAGS: -lmosquitto
#include "mosquitto_ext.h"
*/
import "C"

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"unsafe"
)

func init() {
	C.mosquitto_lib_init()
}

type Conn struct {
	Id       string
	mosq     *C.struct_mosquitto
	handlers map[string]Handler
	wg       sync.WaitGroup
}

func Dial(id string, address string, clean bool) (Conn, error) {
	address_host, address_port, err := net.SplitHostPort(address)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(address_port)
	if err != nil {
		panic(err)
	}

	cid := C.CString(id)
	// TODO: Bug https://code.google.com/p/go/issues/detail?id=4417
	// cclean := C.bool(clean)
	cclean := bool_to_cint(clean)
	chost := C.CString(address_host)
	cport := C.int(port)
	defer C.free(unsafe.Pointer(cid))
	defer C.free(unsafe.Pointer(chost))

	// Setup obj early so mosquitto can pass it around.
	c := Conn{Id: id, handlers: make(map[string]Handler)}

	// TODO: Keepalive.
	// TODO: Bug https://code.google.com/p/go/issues/detail?id=4417
	// c.mosq = C.mosquitto_new(cid, cclean, unsafe.Pointer(&c))
	c.mosq = C.mosquitto_new2(cid, cclean, unsafe.Pointer(&c))
	if c.mosq == nil {
		err = ccode_to_error(C.MOSQ_ERR_ERRNO)
		return c, err
	}

	err = ccode_to_error(C.mosquitto_connect(c.mosq, chost, cport, C.int(60)))
	return c, err
}

func (c *Conn) HandleFunc(sub string, qos int, hf HandlerFunc) error {
	handler, err := NewHandler(sub, qos, hf)
	if err == nil {
		c.handlers[sub] = handler
	}
	return err
}

func (c *Conn) Close() error {
	c.wg.Wait()
	C.mosquitto_disconnect(c.mosq)
	C.mosquitto_destroy(c.mosq)
	return nil
}

// TODO: Message.Id
func (c *Conn) Publish(m Message) error {
	ctopic := C.CString(m.Topic)
	cpayload := unsafe.Pointer(&m.Payload[0])
	cpayloadlen := C.int(len(m.Payload))
	cqos := C.int(m.Qos)
	// TODO: Bug https://code.google.com/p/go/issues/detail?id=4417
	// cretain := C.bool(m.Retain)
	cretain := bool_to_cint(m.Retain)

	defer C.free(unsafe.Pointer(ctopic))
	// TODO: Bug https://code.google.com/p/go/issues/detail?id=4417
	// return code_to_error(C.mosquitto_publish(c.mosq, nil, ctopic, cpayloadlen, cpayload, cqos, cretain))
	return ccode_to_error(C.mosquitto_publish2(c.mosq, nil, ctopic, cpayloadlen, cpayload, cqos, cretain))
}

func (c *Conn) Listen() error {
	return ccode_to_error(C.mosquitto_loop_forever(c.mosq, C.int(-1), C.int(1)))
}

//export on_connect
func on_connect(cconn unsafe.Pointer) {
	c := (*Conn)(cconn)

	// Setup handlers again.
	for _, handler := range c.handlers {
		cqos := C.int(handler.Qos)
		csub := C.CString(handler.Sub)
		C.mosquitto_subscribe(c.mosq, nil, csub, cqos)
		C.free(unsafe.Pointer(csub))
	}
}

//export on_message
func on_message(cconn unsafe.Pointer, ctopic *C.char, cpayload unsafe.Pointer, cpayloadlen C.int) {
	c := (*Conn)(cconn)
	topic := C.GoString(ctopic)

	message, err := NewMessage(topic, C.GoBytes(cpayload, cpayloadlen))
	if err != nil {
		// TODO: Log error.
		return
	}

	for _, handler := range c.handlers {
		if !handler.Match(topic) {
			continue
		}

		c.wg.Add(1)
		go func(h Handler, c *Conn, m Message) {
			h.Call(c, m)
			c.wg.Done()
		}(handler, c, message)
	}
}

//export on_log
func on_log(cconn unsafe.Pointer, clevel C.int, cmessage *C.char) {
	// TODO: Logging.
	// c := (*Conn)(cconn)
	fmt.Printf("%s\n", C.GoString(cmessage))
}
