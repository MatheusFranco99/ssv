package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCSend(signedMessage *messages.SignedMessage) error {

	// Decode
	vcbcSendData, err := signedMessage.Message.GetVCBCSendData()
	if err != nil {
		errors.New("uponVCBCSend: could not get vcbcSendData data from signedMessage")
	}

	// Get sender
	sender := signedMessage.GetSigners()[0]
	data := vcbcSendData.Data

	// Funciton identifier
	i.State.VCBCSendLogTag += 1

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$"+cYellow+" UponVCBCSend "+reset+fmt.Sprint(i.State.VCBCSendLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(sender)), zap.Int("own operator id", int(i.State.Share.OperatorID)))
	}

	log("start")

	// Init time if not initialized before
	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	// Check if already send VCBC Ready
	hasSent := i.State.SentReadys.Has(sender)

	// If never sent ready, or have already sent and the data is equal
	if !hasSent || (hasSent && i.State.SentReadys.EqualData(sender, data)) {

		// Update state
		i.State.SentReadys.Add(sender, data)

		// Create VCBCReady message with hash
		hash, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(data)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
		if err != nil {
			return errors.Wrap(err, "uponVCBCSend: could not compute data hash")
		}

		vcbcReadyMsg, err := i.CreateVCBCReady(hash, sender)
		if err != nil {
			return errors.New("uponVCBCSend: failed to create VCBCReady message with proof")
		}

		i.Broadcast(vcbcReadyMsg)
		log(fmt.Sprintf("%v sent ready to %v", int(i.State.Share.OperatorID), int(sender)))
	}

	return nil
}

func isValidVCBCSend(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
	logger *zap.Logger,
) error {

	// logger
	log := func(str string) {

		if state.HideLogs || state.HideValidationLogs || state.DecidedLogOnly {
			return
		}
		logger.Debug("$$$$$$ UponMV_VCBCSend : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if signedMsg.Message.MsgType != messages.VCBCSendMsgType {
		return errors.New("msg type is not VCBCSend")
	}
	log("checked message type")
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	log("checked message height")
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	log("checked number of signers is 1")

	VCBCSendData, err := signedMsg.Message.GetVCBCSendData()

	log("got vcbc send data")
	if err != nil {
		return errors.Wrap(err, "could not get vcbcsend data")
	}
	if err := VCBCSendData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCSendData invalid")
	}
	log("validated")

	return nil
}

func (i *Instance) CreateVCBCSend(data []byte) (*messages.SignedMessage, error) {

	state := i.State

	vcbcSendData := &messages.VCBCSendData{
		Data: data,
	}
	dataByts, err := vcbcSendData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCSend: could not encode vcbcSendData")
	}
	msg := &messages.Message{
		MsgType:    messages.VCBCSendMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig, hash_map, err := i.Sign(msg)
	if err != nil {
		panic(err)
	}

	signedMsg := &messages.SignedMessage{
		Signature:          sig,
		Signers:            []types.OperatorID{state.Share.OperatorID},
		Message:            msg,
		DiffieHellmanProof: hash_map,
	}
	return signedMsg, nil
}
