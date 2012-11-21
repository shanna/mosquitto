package mosquitto

/*
#cgo LDFLAGS: -lmosquitto

#include <stdlib.h>
#include <stdio.h>
#include <mosquitto.h>

#define GO_MOSQ_DEBUG 0

extern void on_message(void *conn, char* topic, void* payload, int payloadlen);

static void m_message_callback(struct mosquitto *mosq, void *obj, const struct mosquitto_message *message) {
  on_message(obj, message->topic, message->payload, message->payloadlen);
}

static void m_log_callback(struct mosquitto *mosq, void *obj, int level, const char *str) {
#ifdef GO_MOSQ_DEBUG
  printf("%s\n", str);
#endif
}

static void set_callbacks(struct mosquitto *mosq) {
  mosquitto_log_callback_set(mosq, m_log_callback);
  mosquitto_message_callback_set(mosq, m_message_callback);
  // TODO: Other callbacks.
}

// Wrappers for methods with boolean values. See bug: https://code.google.com/p/go/issues/detail?id=4417

static int mosquitto_publish2(struct mosquitto *mosq, int *mid, const char *topic, int payloadlen, const void *payload, int qos, int retain) {
  return mosquitto_publish(mosq, mid, topic, payloadlen, payload, qos, retain);
}

static struct mosquitto *mosquitto_new2(const char *id, int clean_session, void *obj) {
  return mosquitto_new(id, clean_session, obj);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"unsafe"
)

//export on_message
func on_message(conn unsafe.Pointer, topic *C.char, payload unsafe.Pointer, payloadlen C.int) {
	c := (*Conn)(conn)
	message, err := NewMessage(C.GoString(topic), C.GoBytes(payload, payloadlen))
	if err != nil {
		// TODO: Log error.
		return
	}

	// Find handler for message.
	for sub, handler := range c.handlers {
		csub := C.CString(sub)
		cresult := C.bool(false)
		defer C.free(unsafe.Pointer(csub))

		C.mosquitto_topic_matches_sub(csub, topic, &cresult)
		if cresult {
			(*handler)(c, message)
		}
	}
}

type HandlerFunc func(*Conn, Message)

type Message struct {
	Id      int
	Topic   string
	Payload []byte
	Qos     int
	Retain  bool
}

func NewMessage(topic string, payload []byte) (Message, error) {
	return Message{Topic: topic, Payload: payload, Qos: 0, Retain: false}, nil
}

type Conn struct {
	Id       string
	mosq     *C.struct_mosquitto
	handlers map[string]*HandlerFunc
}

func init() {
	C.mosquitto_lib_init()
}

// TODO: I don't understand what the ID is for yet.
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
	c := Conn{Id: id, handlers: make(map[string]*HandlerFunc)}

	// TODO: Keepalive.
	// TODO: Bug https://code.google.com/p/go/issues/detail?id=4417
	// c.mosq = C.mosquitto_new(cid, cclean, unsafe.Pointer(&c))
	c.mosq = C.mosquitto_new2(cid, cclean, unsafe.Pointer(&c))
	if err = code_to_error(C.mosquitto_connect(c.mosq, chost, cport, 60)); err != nil {
		return Conn{}, err
	}

	// Setup C callbacks.
	C.set_callbacks(c.mosq)
	return c, nil
}

func (c *Conn) HandleFunc(sub string, qos int, handler HandlerFunc) error {
	cqos := C.int(qos)
	csub := C.CString(sub)
	defer C.free(unsafe.Pointer(csub))

	// TODO: I don't understand the second argument yet.
	if err := code_to_error(C.mosquitto_subscribe(c.mosq, nil, csub, cqos)); err != nil {
		return err
	}

	// Add handler.
	c.handlers[sub] = &handler
	return nil
}

func (c *Conn) Close() error {
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
	return code_to_error(C.mosquitto_publish2(c.mosq, nil, ctopic, cpayloadlen, cpayload, cqos, cretain))
}

// Channels! This just blocks for now.
func (c *Conn) Listen() {
	// TODO: Timeout.
	for {
		if err := code_to_error(C.mosquitto_loop(c.mosq, C.int(-1), C.int(1))); err != nil {
			fmt.Printf("error: %s\n", err)
			break
		}
	}
}

func bool_to_cint(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func code_to_error(code C.int) error {
	switch code {
	default:
		return errors.New(fmt.Sprintf("Unknown error code: %d", code))
	case C.MOSQ_ERR_SUCCESS:
		return nil
	case C.MOSQ_ERR_INVAL:
		return errors.New("The input parameters were invalid.")
	case C.MOSQ_ERR_NOMEM:
		panic("An out of memory condition occurred.")
	case C.MOSQ_ERR_NO_CONN:
		return errors.New("The client isn't connected to a broker.")
	case C.MOSQ_ERR_CONN_LOST:
		return errors.New("The connection to the broker was lost.")
	case C.MOSQ_ERR_PROTOCOL:
		return errors.New("There is a protocol error communicating with the broker.")
	case C.MOSQ_ERR_ERRNO:
		return errors.New("System call returned an error.")
		// TODO: If a system call returned an error. The variable errno contains the error code, even on Windows.
		// Use strerror_r() where available or FormatMessage() on Windows.
	case C.MOSQ_ERR_PAYLOAD_SIZE:
		return errors.New("Payload is too large.")
	}
	return nil
}
