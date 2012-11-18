package main

import (
  "bitbucket.org/shanehanna/mosquitto"
  "fmt"
)

func main() {
  conn, err := mosquitto.Dial("30", "localhost:1883", true)
  if err != nil { panic(err) }
  fmt.Printf("%+V", conn)
}
