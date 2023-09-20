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

func (m *MessageContainer) AddMessage(s *SignedMessage) {
	m.Msgs = append(m.Msgs, s)
}

func (m *MessageContainer) GetMessagesSlice(a int, b int) []*SignedMessage {
	return m.Msgs[a : b+1]
}
