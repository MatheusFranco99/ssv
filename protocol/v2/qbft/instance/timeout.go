package instance

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)


func (i *Instance) UponRoundTimeout() error {

	round := int(i.State.Round)

	i.logger.Debug("$$$$$$ UponRoundTimeout start. time(micro):",zap.Int64("time(micro)",makeTimestamp()),zap.Int("round",round))

	newRound := i.State.Round + 1
	i.logger.Debug("round timed out", zap.Uint64("round", uint64(newRound)))
	i.State.Round = newRound
	i.State.ProposalAcceptedForCurrentRound = nil
	i.config.GetTimer().TimeoutForRound(i.State.Round)

	roundChange, err := CreateRoundChange(i.State, i.config, newRound, i.StartValue)
	if err != nil {
		return errors.Wrap(err, "could not generate round change msg")
	}
	i.logger.Debug("$$$$$$ UponRoundTimeout broadcast start. time(micro):",zap.Int64("time(micro)",makeTimestamp()),zap.Int("round",round))

	if err := i.Broadcast(roundChange); err != nil {
		return errors.Wrap(err, "failed to broadcast round change message")
	}
	i.logger.Debug("$$$$$$ UponRoundTimeout broadcast finish. time(micro):",zap.Int64("time(micro)",makeTimestamp()),zap.Int("round",round))

	i.logger.Debug("$$$$$$ UponRoundTimeout return. time(micro):",zap.Int64("time(micro)",makeTimestamp()),zap.Int("round",round))

	return nil
}
