package mosquitto

import (
	"fmt"
	"github.com/bmizerany/assert"
	"testing"
  "os"
)

func TestMosquitto(t *testing.T) {
	// Connect.
	conn, err := Dial("tests", "localhost:1883", true)
	assert.Equal(t, nil, err)
	defer conn.Close()

	// Subscribe.
	err = conn.HandleFunc("foo", 2, func(c *Conn, m Message) {
		fmt.Fprintf(os.Stderr, "message: %+v\n", m)
		// TODO: Test we actually got a message here.
		c.Close() // We are done.
	})
	assert.Equal(t, nil, err)

	// Message.
	message, err := NewMessage("foo", []byte("hello world"))
	assert.Equal(t, nil, err)

	// Publish.
	err = conn.Publish(message)
	assert.Equal(t, nil, err)

	// Listen.
	// conn.Listen()
}
