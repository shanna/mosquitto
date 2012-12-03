package main

import (
	"bitbucket.org/shanehanna/mosquitto"
	"fmt"
)

func main() {
	conn, err := mosquitto.Dial("example", "localhost:1883", true)
	if err != nil {
		panic(err)
	}

	err = conn.HandleFunc("foo", 2, func(c *mosquitto.Conn, m mosquitto.Message) {
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

	conn.HandleFunc("bar", 2, func(c *mosquitto.Conn, m mosquitto.Message) {
		fmt.Printf("bar <- (%s)\n", m.Payload)
		c.Close()
	})
	if err != nil {
		panic(err)
	}

	message, _ := mosquitto.NewMessage("foo", []byte("hello world"))
	fmt.Printf("(%s) -> foo\n", message.Payload)
	if err = conn.Publish(message); err != nil {
		panic(err)
	}
	conn.Listen()
}
