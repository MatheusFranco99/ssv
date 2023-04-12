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
	author := vcbcSendData.Author
	priority := vcbcSendData.Priority
	data := vcbcSendData.Data

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCSend "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(sender)))
	}

	log("start")

	log(fmt.Sprintf("i.initTime status: %v", i.initTime))
	if i.initTime == -1 || i.initTime == 0 {
		i.initTime = makeTimestamp()
	}
	log(fmt.Sprintf("i.initTime new: %v", i.initTime))

	// check if it was already received. If not -> store
	log(fmt.Sprintf("has author, priority: %v", i.State.VCBCState.Has(author, priority)))
	if !i.State.VCBCState.Has(author, priority) {
		log("add")

		i.State.VCBCState.Add(author, priority, data)
	}

	if sender != author {
		log("sender != author, quitting.")
		return nil
	}

	log("get hash")
	hash, err := GetDataHash(data)
	if err != nil {
		return errors.New("uponVCBCSend: could not get hash of proposals")
	}

	// create VCBCReady message with proof
	log("create vcbc ready")

	vcbcReadyMsg, err := CreateVCBCReady(i.State, i.config, hash, priority, author)
	if err != nil {
		return errors.New("uponVCBCSend: failed to create VCBCReady message with proof")
	}
	// FIX ME : send specifically to author
	log("broadcast start")
	i.Broadcast(vcbcReadyMsg)
	log("broadcast finish")

	log("finish")

	return nil
}

func isValidVCBCSend(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.VCBCSendMsgType {
	// 	return errors.New("msg type is not VCBCSend")
	// }
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	VCBCSendData, err := signedMsg.Message.GetVCBCSendData()
	if err != nil {
		return errors.Wrap(err, "could not get vcbcsend data")
	}
	if err := VCBCSendData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCSendData invalid")
	}

	// author
	author := VCBCSendData.Author
	authorInCommittee := false
	for _, opID := range operators {
		if opID.OperatorID == author {
			authorInCommittee = true
		}
	}
	if !authorInCommittee {
		return errors.New("author (OperatorID) doesn't exist in Committee")
	}

	if author != signedMsg.GetSigners()[0] {
		return errors.New("author of VCBCSend differs from sender of the message")
	}

	// priority
	// priority := VCBCSendData.Priority
	// if state.VCBCState.HasM(author, priority) {
	// 	if !state.VCBCState.EqualM(author, priority, VCBCSendData.Proposals) {
	// 		return errors.New("existing (priority,author) with different proposals")
	// 	}
	// }

	return nil
}

func CreateVCBCSend(state *messages.State, config alea.IConfig, data []byte, priority specalea.Priority, author types.OperatorID) (*messages.SignedMessage, error) {
	vcbcSendData := &messages.VCBCSendData{
		Data:     data,
		Priority: priority,
		Author:   author,
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
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCSend: failed signing filler msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
