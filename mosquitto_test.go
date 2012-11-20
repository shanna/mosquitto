package mosquitto

import (
  "github.com/bmizerany/assert"
  "testing"
  "fmt"
)

func TestMosquitto(t *testing.T) {
  // Connect.
  conn, err := Dial("30", "localhost:1883", true)
  assert.Equal(t, nil, err)
  defer conn.Close()

  // Subscribe.
  err = conn.Handle("$SYS/#", 3, func(m *Message) {
    fmt.Printf("%+v\n", m)
  })
  assert.Equal(t, nil, err)

  conn.Listen()
}
