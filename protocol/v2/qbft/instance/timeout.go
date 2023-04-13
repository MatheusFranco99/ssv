package instance

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) UponRoundTimeout() error {

	round := int(i.State.Round)

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponRoundTimeout "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("round", int(round)))
	}

	log("start")

	// i.logger.Debug("$$$$$$ UponRoundTimeout start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("round", round))

	newRound := i.State.Round + 1
	// i.logger.Debug("round timed out", zap.Uint64("round", uint64(newRound)))
	i.State.Round = newRound
	i.State.ProposalAcceptedForCurrentRound = nil
	i.config.GetTimer().TimeoutForRound(i.State.Round)
	log("create round change msg")

	roundChange, err := CreateRoundChange(i.State, i.config, newRound, i.StartValue)
	if err != nil {
		return errors.Wrap(err, "could not generate round change msg")
	}
	// i.logger.Debug("$$$$$$ UponRoundTimeout broadcast start. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("round", round))
	log("broadcast start")

	if err := i.Broadcast(roundChange); err != nil {
		return errors.Wrap(err, "failed to broadcast round change message")
	}
	log("broadcast finish")
	log("finish")

	// i.logger.Debug("$$$$$$ UponRoundTimeout broadcast finish. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("round", round))

	// i.logger.Debug("$$$$$$ UponRoundTimeout return. time(micro):", zap.Int64("time(micro)", makeTimestamp()), zap.Int("round", round))

	return nil
}
