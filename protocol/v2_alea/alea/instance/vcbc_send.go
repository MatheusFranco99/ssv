package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCSend(signedMessage *messages.SignedMessage) error {

	// get Data
	vcbcSendData, err := signedMessage.Message.GetVCBCSendData()
	if err != nil {
		errors.New("uponVCBCSend: could not get vcbcSendData data from signedMessage")
	}

	// sender
	sender := signedMessage.GetSigners()[0]
	data := vcbcSendData.Data

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (i.State.DecidedLogOnly) {
			return
		}
		i.logger.Debug("$$$$$$ UponVCBCSend "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(sender)))
	}

	log("start")


	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	has_sent := i.State.SentReadys.Has(sender)
	log(fmt.Sprintf("check if has sent %v",has_sent))

	// if never sent ready, or have already sent and the data is equal
	if (!has_sent || (has_sent && i.State.SentReadys.EqualData(sender,data))) {


		i.State.SentReadys.Add(sender,data)
		log("added to sent readys structure")

		// create VCBCReady message with hash
		hash, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(data)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
		if err != nil {
			return errors.Wrap(err, "uponVCBCSend: could not compute data hash")
		}
		log("computed hash")


		vcbcReadyMsg, err := CreateVCBCReady(i.State, i.config, hash, sender)
		if err != nil {
			return errors.New("uponVCBCSend: failed to create VCBCReady message with proof")
		}
		log("created VCBCReady")

		// FIX ME : send specifically to author
		i.Broadcast(vcbcReadyMsg)
		log("broadcasted")
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

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (state.DecidedLogOnly) {
			return
		}
		logger.Debug("$$$$$$ UponMV_VCBCSend "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
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
	// if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
	// 	return errors.Wrap(err, "msg signature invalid")
	// }

	// log("checked signature")

	msg_bytes, err := signedMsg.Message.Encode()
	if err != nil {
		return errors.Wrap(err, "Could not encode message")
	}
	if !state.DiffieHellmanContainerOneTimeCost.VerifyHash(msg_bytes,signedMsg.GetSigners()[0],signedMsg.DiffieHellmanProof[state.Share.OperatorID]) {
		return errors.New("Failed Diffie Hellman verification")
	}

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

func CreateVCBCSend(state *messages.State, config alea.IConfig, data []byte) (*messages.SignedMessage, error) {
	vcbcSendData := &messages.VCBCSendData{
		Data:     data,
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

	// No signing -> use Diffie Hellman
	// sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "CreateVCBCSend: failed signing filler msg")
	// }
	sig := []byte{}

	msg_bytes, err := msg.Encode()
	if err != nil {
		return nil, errors.Wrap(err,"CreateVCBCSend: failed to encode message")
	}
	hash_map := state.DiffieHellmanContainerOneTimeCost.GetHashMap(msg_bytes)

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
		DiffieHellmanProof: hash_map,
	}
	return signedMsg, nil
}
