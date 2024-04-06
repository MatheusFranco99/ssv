package validator

import (
	"fmt"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/queue"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/runner"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (v *Validator) handleEventMessage(msg *queue.DecodedSSVMessage, dutyRunner runner.Runner) error {
	eventMsg, ok := msg.Body.(*types.EventMsg)
	if !ok {
		return errors.New("could not decode event message")
	}
	switch eventMsg.Type {
	case types.Timeout:
		err := dutyRunner.GetBaseRunner().QBFTController.OnTimeout(*eventMsg)
		if err != nil {
			v.logger.Warn("on timeout failed", zap.Error(err)) // need to return error instead?
		}
		return nil
	case types.ExecuteDuty:
		err := v.OnExecuteDuty(*eventMsg)
		if err != nil {
			v.logger.Warn("failed to execute duty", zap.Error(err)) // need to return error instead?
		}
		return nil
	default:
		return errors.New(fmt.Sprintf("unknown event msg - %s", eventMsg.Type.String()))
	}
}
