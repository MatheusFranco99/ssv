package instance

import (
	"fmt"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

// ProcessMsg processes a new QBFT msg, returns non nil error on msg processing error
func (i *Instance) ProcessMsg(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	// If not started -> quit
	if !i.State.HasStarted {
		return false, nil, nil, nil
	}

	// Ready Special treatment: if you're not the author, don't process it
	// this is occurs due to the pubsub topology
	if should_stop, err := i.CheckReadyFiter(msg); should_stop {
		return false, nil, nil, err
	}

	return i.ValidateAndProcess(msg)
}

func (i *Instance) ValidateAndProcess(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	// Validate fields formatting
	if err := msg.Validate(); err != nil {
		return false, nil, nil, errors.Wrap(err, "invalid signed message")
	}

	if msg.Message.Round < i.State.Round {
		return false, nil, nil, errors.New("past round")
	}

	// Normal BLS
	if i.State.UseBLS && !i.State.AggregateVerify {
		return i.BLSBehaviorProcessing(msg)
	} else if i.State.AggregateVerify {
		return i.BLSAggBehaviorProcessing(msg)
	} else if i.State.UseDiffieHellman {
		return i.HMACBehaviorProcessing(msg)
	} else if i.State.UseRSA {
		return i.RSABehaviorProcessing(msg)
	} else if i.State.UseEDDSA {
		return i.EDDSABehaviorProcessing(msg)
	} else {
		panic("No Signature behavior set.")
	}
}

// Ready Special treatment: returns true if the msg is a VCBC Ready and you're not the author or you have already sent VCBC Final
func (i *Instance) CheckReadyFiter(msg *messages.SignedMessage) (bool, error) {

	if msg.Message.MsgType == messages.VCBCReadyMsgType {

		// Decode
		vcbcReadyData, err := msg.Message.GetVCBCReadyData()
		if err != nil {
			return true, errors.Wrap(err, "ProcessMsg: could not get vcbcReadyData data from signedMessage")
		}

		// Get attributes
		author := vcbcReadyData.Author
		if author != i.State.Share.OperatorID {
			return true, nil
		}

		// If already sent final, don't process more readys
		if i.State.ReceivedReadys.HasSentFinal() {
			return true, nil
		}
	}
	return false, nil
}

func (i *Instance) ValidateSemantics(msg *messages.SignedMessage) error {
	// Validate Semantics for each message
	var err error
	switch msg.Message.MsgType {
	case messages.VCBCSendMsgType:
		err = isValidVCBCSend(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.VCBCReadyMsgType:
		err = isValidVCBCReady(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.VCBCFinalMsgType:
		err = isValidVCBCFinal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAInitMsgType:
		err = isValidABAInit(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAAuxMsgType:
		err = isValidABAAux(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAConfMsgType:
		err = isValidABAConf(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAFinishMsgType:
		err = isValidABAFinish(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.CommonCoinMsgType:
		err = isValidCommonCoin(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	default:
		err = errors.New(fmt.Sprintf("signed message type not supported: %v", msg.Message.MsgType))
	}

	return err
}

func (i *Instance) ProcessMsgLogic(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {

	decided = false
	decidedValue = nil
	aggregatedCommit = nil
	res := i.processMsgF.Run(func() interface{} {
		switch msg.Message.MsgType {
		case messages.ABAInitMsgType:
			return i.uponABAInit(msg)
		case messages.ABAAuxMsgType:
			return i.uponABAAux(msg)
		case messages.ABAConfMsgType:
			return i.uponABAConf(msg)
		case messages.ABAFinishMsgType:
			return i.uponABAFinish(msg)
		case messages.VCBCSendMsgType:
			return i.uponVCBCSend(msg)
		case messages.VCBCReadyMsgType:
			return i.uponVCBCReady(msg)
		case messages.VCBCFinalMsgType:
			return i.uponVCBCFinal(msg)
		case messages.CommonCoinMsgType:
			return i.uponCommonCoin(msg)
		default:
			return errors.New("signed message type not supported")
		}
	})
	if res != nil {
		return false, nil, nil, res.(error)
	}

	decided = i.State.Decided
	decidedValue = i.State.DecidedValue
	aggregatedCommit = i.State.DecidedMessage
	return i.State.Decided, i.State.DecidedValue, i.State.DecidedMessage, nil
}

func (i *Instance) ProcessBufferOfMessages(msgs []*messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {
	decided = false
	decidedValue = nil
	aggregatedCommit = nil
	err = nil
	for _, m := range msgs {
		decidedM, decidedValueM, aggregatedCommitM, errM := i.ProcessMsgLogic(m)

		decided = decided || decidedM
		if decidedValue == nil {
			decidedValue = decidedValueM
		}
		if aggregatedCommit == nil {
			aggregatedCommit = aggregatedCommitM
		}
		if err == nil {
			err = errM
		}
	}
	return decided, decidedValue, aggregatedCommit, err
}
