package mosquitto

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

func TestMosquitto(t *testing.T) {
	// Connect.
	conn, err := Dial("tests", "localhost:1883", true)
	if err != nil {
		t.Logf("Dial Error: %s\n", err.Error())
	}
	defer conn.Close()

	// Listen.
	go conn.Listen()

	// Subscribe.
	result := make(chan string, 1)
	err = conn.HandleFunc("foo", 2, func(c *Conn, m Message) {
		result <- string(m.Payload)
	})
	assert.Equal(t, nil, err)

	// Give listener/handler a little time to spin up.
	time.Sleep(1 * time.Second)

	// Message.
	message, err := NewMessage("foo", []byte("hello world"))
	assert.Equal(t, nil, err)

	// Publish.
	err = conn.Publish(message)
	assert.Equal(t, nil, err)

	assert.Equal(t, "hello world", <-result)
}
