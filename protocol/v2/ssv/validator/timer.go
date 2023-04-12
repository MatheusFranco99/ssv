package validator

import (
	"encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/message"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/queue"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (v *Validator) onTimeout(identifier spectypes.MessageID, height specalea.Height) func() {
	return func() {
		dr := v.DutyRunners[identifier.GetRoleType()]
		hasDuty := dr.HasRunningDuty()
		if !hasDuty {
			return
		}

		msg, err := v.createTimerMessage(identifier, height)
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

func (v *Validator) createTimerMessage(identifier spectypes.MessageID, height specalea.Height) (*spectypes.SSVMessage, error) {
	td := types.TimeoutData{Height: height}
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
