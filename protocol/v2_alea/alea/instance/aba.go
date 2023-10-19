package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	// "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

func (i *Instance) StartABA() error {

	//funciton identifier
	i.State.AbaLogTag += 1

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$ UponStartAlea "+fmt.Sprint(i.State.AbaLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("ACRound", int(i.State.ACState.ACRound)))
	}

	log("start")

	if i.State.ACState.IsTerminated() {
		log("ac terminated. quitting.")
		return nil
	}

	acround := i.State.ACState.ACRound
	leader := i.State.Share.Committee[int(acround)%len(i.State.Share.Committee)].OperatorID
	// opIDList := make([]types.OperatorID, len(i.State.Share.Committee))
	// for idx, op := range i.State.Share.Committee {
	// 	opIDList[idx] = op.OperatorID
	// }
	// leader := opIDList[(acround)%len(opIDList)]
	// i.config.GetProposerF()(i.State, specalea.Round(i.State.ACState.ACRound))
	log(fmt.Sprintf("leader: %v", int(leader)))

	vote := byte(0)
	if i.State.VCBCState.HasData(leader) {
		vote = byte(1)
	}
	log(fmt.Sprintf("vote: %v", int(vote)))

	initMsg, err := i.CreateABAInit(vote, specalea.FirstRound, acround)
	if err != nil {
		return errors.Wrap(err, "UponStartABA: failed to create ABA Init message")
	}
	log("created aba init")

	aba := i.State.ACState.GetABA(acround)
	abaround := aba.GetABARound(specalea.FirstRound)
	abaround.SetSentInit(vote)
	log("set sent init")

	i.Broadcast(initMsg)
	log("broadcasted aba init")

	// specialVoteMsg, err := CreateABASpecialVote(i.State, i.config, vote, acround)
	// if err != nil {
	// 	return errors.Wrap(err, "UponStartABA: failed to create ABA special vote message")
	// }
	// log("created aba special vote")

	// i.Broadcast(specialVoteMsg)
	// log("broadcasted aba special vote")

	return nil
}
