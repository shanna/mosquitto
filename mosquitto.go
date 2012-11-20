package mosquitto

/*
#cgo LDFLAGS: -lmosquitto

#include <stdlib.h>
#include <stdio.h>
#include <mosquitto.h>

extern void on_message(void *conn, char* topic, void* payload, int payloadlen);

static void m_message_callback(struct mosquitto *mosq, void *obj, const struct mosquitto_message *message) {
  on_message(obj, message->topic, message->payload, message->payloadlen);
}

static void m_log_callback(struct mosquitto *mosq, void *obj, int level, const char *str) {
  // XXX: Debugging.
  // printf("%s\n", str);
}

static void set_callbacks(struct mosquitto *mosq) {
  mosquitto_log_callback_set(mosq, m_log_callback);
  mosquitto_message_callback_set(mosq, m_message_callback);
  // TODO: Other callbacks.
}

*/
import "C"

import (
	"errors"
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
	chost := C.CString(address_host)
	cport := C.int(port)
	defer C.free(unsafe.Pointer(cid))
	defer C.free(unsafe.Pointer(chost))

	// Setup obj early so mosquitto can pass it around.
	c := Conn{Id: id, handlers: make(map[string]*HandlerFunc)}

	// TODO: Keepalive.
	c.mosq = C.mosquitto_new(cid, C.bool(clean), unsafe.Pointer(&c))
	if C.mosquitto_connect(c.mosq, chost, cport, 60) != 0 {
		return Conn{}, errors.New("Unable to connect.")
	}

	C.set_callbacks(c.mosq)
	return c, nil
}

func (c *Conn) HandleFunc(sub string, qos int, handler HandlerFunc) error {
	cqos := C.int(qos)
	csub := C.CString(sub)
	defer C.free(unsafe.Pointer(csub))

	// TODO: I don't understand the second argument yet.
	if C.mosquitto_subscribe(c.mosq, nil, csub, cqos) != 0 {
		return errors.New("Connection failed.")
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
	cretain := C.bool(m.Retain)
	defer C.free(unsafe.Pointer(ctopic))

	switch r := C.mosquitto_publish(c.mosq, nil, ctopic, cpayloadlen, cpayload, cqos, cretain); r {
	case C.MOSQ_ERR_SUCCESS:
	case C.MOSQ_ERR_INVAL:
		return errors.New("The input parameters were invalid.")
	case C.MOSQ_ERR_NOMEM:
		panic("An out of memory condition occurred.")
	case C.MOSQ_ERR_NO_CONN:
		return errors.New("The client isn't connected to a broker.")
	case C.MOSQ_ERR_PROTOCOL:
		return errors.New("There is a protocol error communicating with the broker.")
	case C.MOSQ_ERR_PAYLOAD_SIZE:
		return errors.New("Payload is too large.")
	}
	return nil
}

// Channels! This just blocks for now.
func (c *Conn) Listen() {
	// TODO: Timeout.
	for {
		C.mosquitto_loop(c.mosq, C.int(-1), C.int(1))
		// Noop.
	}
}
