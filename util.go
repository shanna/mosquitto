package mosquitto

/*
#cgo LDFLAGS: -lmosquitto
#include "mosquitto_ext.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func bool_to_cint(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func ccode_to_error(code C.int) error {
	switch code {
	default:
		return fmt.Errorf("Unknown error code: %d", code)
	case C.MOSQ_ERR_SUCCESS:
		return nil
	case C.MOSQ_ERR_INVAL:
		return Error{(int)(code), "The input parameters were invalid."}
	case C.MOSQ_ERR_NOMEM:
		panic("An out of memory condition occurred.")
	case C.MOSQ_ERR_NO_CONN:
		return Error{(int)(code), "The client isn't connected to a broker."}
	case C.MOSQ_ERR_CONN_LOST:
		return Error{(int)(code), "The connection to the broker was lost."}
	case C.MOSQ_ERR_PROTOCOL:
		return Error{(int)(code), "There is a protocol error communicating with the broker."}
	case C.MOSQ_ERR_ERRNO:
		cerr := C.CString("") // TODO: Is this safe?
		defer C.free(unsafe.Pointer(cerr))
		C.mosquitto_error(cerr)
		return Error{(int)(code), C.GoString(cerr)}
	case C.MOSQ_ERR_PAYLOAD_SIZE:
		return Error{(int)(code), "Payload is too large."}
	}
	return nil
}
