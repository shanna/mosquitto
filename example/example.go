package main

import (
	"bitbucket.org/shanehanna/mosquitto"
	"fmt"
  "time"
)

func main() {
	conn, _ := mosquitto.Dial("example", "localhost:1883", true)
	go conn.Listen()

	err := conn.HandleFunc("foo", 2, func(c *mosquitto.Conn, m mosquitto.Message) {
		fmt.Printf("foo <- (%s)\nfoo -> bar(%s)\n", m.Payload, m.Payload)

		// Change the topic and send it again.
		m.Topic = "bar"
		if err := c.Publish(m); err != nil {
			panic(err)
		}
	})
	if err != nil {
		panic(err)
	}

  done := make(chan bool, 1)
	conn.HandleFunc("bar", 2, func(c *mosquitto.Conn, m mosquitto.Message) {
		fmt.Printf("bar <- (%s)\n", m.Payload)
    done <- true
	})
	if err != nil {
		panic(err)
	}

  time.Sleep(1 * time.Second) // Give the listener/handlers a chance to connect and get set up.

	message, _ := mosquitto.NewMessage("foo", []byte("hello world"))
	fmt.Printf("(%s) -> foo\n", message.Payload)
	if err := conn.Publish(message); err != nil {
		panic(err)
	}
  <-done
}
