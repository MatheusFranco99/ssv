package validator

import (
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/network/forks"
	"go.uber.org/zap"
)

const bufSize = 16384

func newMessageRouter(logger *zap.Logger, msgID forks.MsgIDFunc) *messageRouter {
	return &messageRouter{
		logger: logger,
		ch:     make(chan spectypes.SSVMessage, bufSize),
		msgID:  msgID,
	}
}

type messageRouter struct {
	logger *zap.Logger
	ch     chan spectypes.SSVMessage
	msgID  forks.MsgIDFunc
}

func (r *messageRouter) Route(message spectypes.SSVMessage) {
	select {
	case r.ch <- message:
	default:
		r.logger.Warn("message router buffer is full. dropping message")
	}
}

func (r *messageRouter) GetMessageChan() <-chan spectypes.SSVMessage {
	return r.ch
}
