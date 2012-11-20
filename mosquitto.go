package mosquitto

/*
#cgo LDFLAGS: -lmosquitto

#include <stdlib.h>
#include <stdio.h>
#include <mosquitto.h>

extern void on_message(void *conn, char* topic, char* payload);

static void m_message_callback(struct mosquitto *mosq, void *obj, const struct mosquitto_message *message) {
  on_message(obj, message->topic, message->payload);
}

static void m_log_callback(struct mosquitto *mosq, void *obj, int level, const char *str) {
  printf("%s\n", str);
}

static void set_callbacks(struct mosquitto *mosq) {
  mosquitto_log_callback_set(mosq, m_log_callback); // XXX: Debugging.
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
  // "fmt"
)

//export on_message
func on_message(conn unsafe.Pointer, topic *C.char, payload *C.char) {
  c := (*Conn)(conn)
  for sub, handler := range c.handlers {
    csub    := C.CString(sub)
    cresult := C.bool(false)
    defer C.free(unsafe.Pointer(csub))

    C.mosquitto_topic_matches_sub(csub, topic, &cresult)
    if cresult {
      (*handler)(c, &Message{topic: C.GoString(topic), payload: C.GoString(payload)})
    }
  }
}

type HandlerFunc func(*Conn, *Message)

type Message struct {
  topic, payload string
}

type Conn struct {
  id       string
  mosq     *C.struct_mosquitto
  handlers map[string]*HandlerFunc
}

func init() {
  C.mosquitto_lib_init()
}

// TODO: I don't understand what the ID is for yet.
func Dial(id string, address string, clean bool) (Conn, error) {
  address_host, address_port, err := net.SplitHostPort(address)
  if err != nil { panic(err) }

  port, err := strconv.Atoi(address_port)
  if err != nil { panic(err) }

  cid   := C.CString(id)
  chost := C.CString(address_host)
  cport := C.int(port)
  defer C.free(unsafe.Pointer(cid))
  defer C.free(unsafe.Pointer(chost))

  // Setup obj early so mosquitto can pass it around.
  c := Conn{id: id, handlers: make(map[string]*HandlerFunc)}

  // TODO: Keepalive.
  c.mosq = C.mosquitto_new(cid, C.bool(clean), unsafe.Pointer(&c))
  if (C.mosquitto_connect(c.mosq, chost, cport, 60) != 0) {
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
  if (C.mosquitto_subscribe(c.mosq, nil, csub, cqos) != 0) {
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

// Channels! This just blocks for now.
func (c *Conn) Listen() {
  // TODO: Timeout.
  for {
    C.mosquitto_loop(c.mosq, C.int(-1), C.int(1))
    // Noop.
  }
}
