package mosquitto

/*
#cgo LDFLAGS: -lmosquitto

#include <stdlib.h>
#include <mosquitto.h>
*/
import "C"

import (
  "errors"
  "net"
  "strconv"
  "unsafe"
)

type Conn struct {
  mosq *C.struct_mosquitto
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

  // TODO: Keepalive.
  mosq := C.mosquitto_new(cid, C.bool(clean), nil)
  if (C.mosquitto_connect(mosq, chost, cport, 60) != 0) {
    return Conn{}, errors.New("Unable to connect.")
  }
  return Conn{mosq: mosq}, nil
}

func (c *Conn) Subscribe(sub string, qos int) error {
  cqos := C.int(qos)
  csub := C.CString(sub)
  defer C.free(unsafe.Pointer(csub))

  // TODO: I don't understand the second argument yet.
  if (C.mosquitto_subscribe(c.mosq, nil, csub, cqos) != 0) {
    return errors.New("Connection failed.")
  }
  return nil
}

func (c *Conn) Close() error {
  C.mosquitto_destroy(c.mosq)
  return nil
}
