package main

import (
  "bitbucket.org/shanehanna/mosquitto"
  "fmt"
)

func main() {
  conn, err := mosquitto.Dial("example", "localhost:1883", true)
  if err != nil { panic(err) }
  fmt.Printf("%+V", conn)

  err = conn.HandleFunc("test", 2, func(c* mosquitto.Conn, m *mosquitto.Message) {
    fmt.Printf("example: %+v\n", m)
  })

  conn.Listen()
}
