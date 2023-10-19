package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

func InitializeContainerAcRoundAndRound(container map[specalea.ACRound]map[specalea.Round]*messages.MessageContainer, acround specalea.ACRound, round specalea.Round) {
	if _, ok := container[acround]; !ok {
		container[acround] = make(map[specalea.Round]*messages.MessageContainer)
	}
	if _, ok := container[acround][round]; !ok {
		container[acround][round] = messages.NewMsgContainer()
	}
}
func InitializeContainerAcRound(container map[specalea.ACRound]*messages.MessageContainer, acround specalea.ACRound) {
	if _, ok := container[acround]; !ok {
		container[acround] = messages.NewMsgContainer()
	}
}
func (i *Instance) WaitQuorum(container *messages.MessageContainer) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	if container.Len() == int(i.State.Share.Quorum) {
		err = i.VerifyBLSAggregate(container.GetMessages())
		if err != nil {
			return false, nil, nil, err
		}
		return i.ProcessBufferOfMessages(container.GetMessages())
	} else {
		return i.State.Decided, i.State.DecidedValue, i.State.DecidedMessage, nil
	}
}
func (i *Instance) WaitPartialQuorumAndQuorum(container *messages.MessageContainer) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	if container.Len() == int(i.State.Share.PartialQuorum) {
		err = i.VerifyBLSAggregate(container.GetMessagesSlice(0, int(i.State.Share.PartialQuorum)-1))
		if err != nil {
			return false, nil, nil, err
		}
		return i.ProcessBufferOfMessages(container.GetMessagesSlice(0, int(i.State.Share.PartialQuorum)-1))
	} else if container.Len() == int(i.State.Share.Quorum) {
		err = i.VerifyBLSAggregate(container.GetMessagesSlice(int(i.State.Share.PartialQuorum), int(i.State.Share.Quorum)-1))
		if err != nil {
			return false, nil, nil, err
		}
		return i.ProcessBufferOfMessages(container.GetMessagesSlice(int(i.State.Share.PartialQuorum), int(i.State.Share.Quorum)-1))
	} else {
		return i.State.Decided, i.State.DecidedValue, i.State.DecidedMessage, nil
	}
}

func (i *Instance) WaitVCBCFinal(container *messages.MessageContainer, msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	// If Final:
	// 		i) Wait quorum -> aggregate and BLS verify the aggregated msg fields + process buffer o msgs
	//		ii) post quorum -> BLS verify the aggregated msg + process msg logic

	// i) Quorum -> aggregate and BLS verify the aggregated msg fields + process buffer o msgs
	if container.Len() == int(i.State.Share.Quorum) {
		err := i.VerifyBLSAggregateFinals(container.GetMessages())
		if err != nil {
			return false, nil, nil, err
		}
		return i.ProcessBufferOfMessages(container.GetMessages())

		// ii) Post quorum -> BLS verify the aggregated msg + process msg logic
	} else if container.Len() > int(i.State.Share.Quorum) {

		aggregated_msg, err := GetAggregatedMessageFromVCBCFinal(msg)
		if err != nil {
			return false, nil, nil, err
		}

		if err := aggregated_msg.Signature.VerifyByOperators(aggregated_msg, i.config.GetSignatureDomainType(), types.QBFTSignatureType, i.State.Share.Committee); err != nil {
			return false, nil, nil, errors.Wrap(err, "msg signature invalid")
		}

		return i.ProcessMsgLogic(msg)

		// Wait
	} else {
		return i.State.Decided, i.State.DecidedValue, i.State.DecidedMessage, nil
	}
}
