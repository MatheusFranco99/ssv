package messages

import (
	"bytes"
)

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

func (m *MessageContainer) AllFinalEqual() bool {
	first_msg := &VCBCFinalData{}
	err := first_msg.Decode(m.Msgs[0].Message.Data)
	if err != nil {
		panic(err)
	}
	first_hash := first_msg.Hash

	for i, msg := range m.Msgs {
		if (i == 0) {
			continue
		}
		i_msg := &VCBCFinalData{}
		err := i_msg.Decode(msg.Message.Data)
		if err != nil {
			panic(err)
		}
		
		if (!bytes.Equal(i_msg.Hash,first_hash)) {
			return false
		}
	}
	return true
}


