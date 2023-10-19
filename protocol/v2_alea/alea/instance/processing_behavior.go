package instance

import (
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
)

func (i *Instance) BLSBehaviorProcessing(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {
	if msg.Message.MsgType == messages.VCBCFinalMsgType {

		aggregated_msg, err := GetAggregatedMessageFromVCBCFinal(msg)
		if err != nil {
			return false, nil, nil, err
		}

		i.Verify(aggregated_msg)
	} else {
		i.Verify(msg)
	}
	return i.ProcessMsgLogic(msg)
}

func (i *Instance) BLSAggBehaviorProcessing(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	/*
		VCBC Send: verify + process
		VCBC Ready, Conf, CommonCoin: Wait for Quorum
		VCBC Final: wait threshold and then process each single msg
		ABA Init, Aux, Finish: Wait partial quorum and then quorum -> process batch of turns
	*/
	if msg.Message.MsgType == messages.VCBCSendMsgType {
		i.Verify(msg)
		return i.ProcessMsgLogic(msg)
	} else if msg.Message.MsgType == messages.VCBCReadyMsgType || msg.Message.MsgType == messages.CommonCoinMsgType || msg.Message.MsgType == messages.ABAConfMsgType {

		var container *messages.MessageContainer

		if msg.Message.MsgType == messages.VCBCReadyMsgType {
			container = i.State.ReadyContainer
		} else if msg.Message.MsgType == messages.CommonCoinMsgType {
			container = i.State.CommonCoinMsgContainer
		} else if msg.Message.MsgType == messages.ABAConfMsgType {
			conf_container := i.State.AbaConfContainer

			data, err := msg.Message.GetABAConfData()
			if err != nil {
				return false, nil, nil, err
			}
			acround := data.ACRound
			round := data.Round
			InitializeContainerAcRoundAndRound(conf_container, acround, round)

			container = conf_container[acround][round]
		}

		container.AddMessage(msg)

		return i.WaitQuorum(container)

	} else if msg.Message.MsgType == messages.VCBCFinalMsgType {

		i.State.FinalContainer.AddMessage(msg)

		return i.WaitVCBCFinal(i.State.FinalContainer, msg)

	} else if msg.Message.MsgType == messages.ABAInitMsgType || msg.Message.MsgType == messages.ABAAuxMsgType || msg.Message.MsgType == messages.ABAFinishMsgType {
		var container *messages.MessageContainer

		if msg.Message.MsgType == messages.ABAInitMsgType {
			aba_container := i.State.AbaInitContainer
			data, err := msg.Message.GetABAConfData()
			if err != nil {
				return false, nil, nil, err
			}
			acround := data.ACRound
			round := data.Round
			InitializeContainerAcRoundAndRound(aba_container, acround, round)

			container = aba_container[acround][round]

		} else if msg.Message.MsgType == messages.ABAAuxMsgType {
			aba_container := i.State.AbaAuxContainer
			data, err := msg.Message.GetABAConfData()
			if err != nil {
				return false, nil, nil, err
			}
			acround := data.ACRound
			round := data.Round
			InitializeContainerAcRoundAndRound(aba_container, acround, round)

			container = aba_container[acround][round]

		} else if msg.Message.MsgType == messages.ABAFinishMsgType {
			aba_container := i.State.AbaFinishContainer

			data, err := msg.Message.GetABAConfData()
			if err != nil {
				return false, nil, nil, err
			}
			acround := data.ACRound
			InitializeContainerAcRound(aba_container, acround)

			container = aba_container[acround]
		}

		container.AddMessage(msg)

		return i.WaitPartialQuorumAndQuorum(container)

	} else {
		panic("Unknown message type")
	}
}

func (i *Instance) HMACBehaviorProcessing(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	// If Final:
	// 		i) Wait quorum -> aggregate and BLS verify the aggregated msg fields + process buffer o msgs
	//		ii) post quorum -> BLS verify the aggregated msg + process msg logic
	if msg.Message.MsgType == messages.VCBCFinalMsgType {

		i.State.FinalContainer.AddMessage(msg)

		return i.WaitVCBCFinal(i.State.FinalContainer, msg)

		// If Ready -> Wait for quorum and verify aggregated signature
	} else if msg.Message.MsgType == messages.VCBCReadyMsgType {

		i.State.ReadyContainer.AddMessage(msg)

		return i.WaitQuorum(i.State.ReadyContainer)

		// Any other message type -> Verify (hmac) and process
	} else {
		i.Verify(msg)
		return i.ProcessMsgLogic(msg)
	}
}

func (i *Instance) RSABehaviorProcessing(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {
	if msg.Message.MsgType == messages.VCBCFinalMsgType {
		// simulate verifying a buffer of a quorum of messagaes
		num_verifications := uint64(0)
		for num_verifications < i.State.Share.Quorum {
			i.Verify(msg)
			num_verifications += 1
		}
	} else {
		i.Verify(msg)
	}
	return i.ProcessMsgLogic(msg)
}

func (i *Instance) EDDSABehaviorProcessing(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {
	// Does the same thing as RSA
	return i.RSABehaviorProcessing(msg)
}
