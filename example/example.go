package main

import (
	"bitbucket.org/shanehanna/mosquitto"
	"fmt"
  "os"
)

func main() {
	conn, err := mosquitto.Dial("example", "localhost:1883", true)
	if err != nil { panic(err) }

  err = conn.HandleFunc("foo", 2, func(c *mosquitto.Conn, m *mosquitto.Message) {
		fmt.Printf("foo <- (%s)\nfoo -> bar(%s)\n", m.Payload, m.Payload)
    if err := c.Publish("bar", m.Payload); err != nil {
      panic(err)
    }
	})
  if err != nil { panic(err) }
  conn.HandleFunc("bar", 2, func(c *mosquitto.Conn, m *mosquitto.Message) {
    fmt.Printf("bar <- (%s)\n", m.Payload)
    conn.Close()
    os.Exit(0)
  })
  if err != nil { panic(err) }

  conn.Listen()
  /*
  go conn.Listen()

  payload := "hello world"
  fmt.Printf("(%s) -> foo\n", payload)
  conn.Publish("foo", payload)
  for {} // Loop.
  */
}
