package validator

import (
	"encoding/json"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2/message"
	"github.com/MatheusFranco99/ssv/protocol/v2/ssv/queue"
	"github.com/MatheusFranco99/ssv/protocol/v2/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (v *Validator) onTimeout(identifier spectypes.MessageID) func() {
	return func() {
		dr := v.DutyRunners[identifier.GetRoleType()]
		hasDuty := dr.HasRunningDuty()
		if !hasDuty {
			return
		}

		msg, err := v.createTimerMessage(identifier)
		if err != nil {
			v.logger.Debug("failed to create timer msg", zap.Error(err))
			return
		}
		dec, err := queue.DecodeSSVMessage(msg)
		if err != nil {
			v.logger.Debug("failed to decode timer msg", zap.Error(err))
			return
		}
		v.Queues[identifier.GetRoleType()].Q.Push(dec)
	}
}

func (v *Validator) createTimerMessage(identifier spectypes.MessageID) (*spectypes.SSVMessage, error) {
	td := types.TimeoutData{}
	data, err := json.Marshal(td)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal timout data")
	}
	eventMsg := &types.EventMsg{
		Type: types.Timeout,
		Data: data,
	}

	eventMsgData, err := eventMsg.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode timeout signed msg")
	}
	return &spectypes.SSVMessage{
		MsgType: message.SSVEventMsgType,
		MsgID:   identifier,
		Data:    eventMsgData,
	}, nil
}
