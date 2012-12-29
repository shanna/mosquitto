package mosquitto

type Message struct {
	Id      int
	Topic   string
	Payload []byte
	Qos     int
	Retain  bool
}

func NewMessage(topic string, payload []byte) (Message, error) {
	return Message{0, topic, payload, 0, false}, nil
}

