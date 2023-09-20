package messages

type MessageContainer struct {
	Msgs []*SignedMessage
}

func NewMsgContainer() *MessageContainer {
	return &MessageContainer{
		Msgs: []*SignedMessage{},
	}
}

func (m *MessageContainer) Len() int {
	return len(m.Msgs)
}

func (m *MessageContainer) GetMessages() []*SignedMessage {
	return m.Msgs
}
