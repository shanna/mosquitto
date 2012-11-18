package mosquitto

import (
  "github.com/bmizerany/assert"
  "testing"
)

func TestDial(t *testing.T) {
  conn, err := Dial("30", "localhost:1883", true)
  defer conn.Close()

  assert.Equal(t, nil, err)
}

