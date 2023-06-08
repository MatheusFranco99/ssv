package instance

import (
	// "bytes"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCReady(signedMessage *messages.SignedMessage) error {

	// get Data
	vcbcReadyData, err := signedMessage.Message.GetVCBCReadyData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not get vcbcReadyData data from signedMessage")
	}

	// get attributes
	hash := vcbcReadyData.Hash
	author := vcbcReadyData.Author
	if author != i.State.Share.OperatorID {
		return nil
	}
	signature := vcbcReadyData.Signature
	// priority := vcbcReadyData.Priority
	senderID := signedMessage.GetSigners()[0]

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCReady "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("sender", int(senderID)))
	}

	log("start")

	i.State.ReceivedReadys.Add(senderID,signature)
	log(fmt.Sprintf("len, quorum: %v %v", i.State.ReceivedReadys.GetLen(), int(i.State.Share.Quorum)))


	already_has_quorum := (i.State.ReceivedReadys.GetLen() >= int(i.State.Share.Quorum))

	log(fmt.Sprintf("already has quorum %v", already_has_quorum))
	log(fmt.Sprintf("Has sent final %v",i.State.ReceivedReadys.HasSentFinal()))

	if (already_has_quorum && !i.State.ReceivedReadys.HasSentFinal())  {

		log("creating vcbc final")
		vcbcFinalMsg, err := CreateVCBCFinal(i.State, i.config, i.StartValue, hash, author, []byte{}, i.State.ReceivedReadys.GetNodeIDs())
		if err != nil {
			return errors.Wrap(err, "uponVCBCReady: failed to create VCBCReady message with proof")
		}

		log("broadcast start")
		i.Broadcast(vcbcFinalMsg)
		log("broadcast finish")
	}

	log("finish")

	return nil
}

func AggregateMsgs(msgs []*messages.SignedMessage) (*messages.SignedMessage, error) {
	if len(msgs) == 0 {
		return nil, errors.New("AggregateMsgs: can't aggregate zero msgs")
	}

	var ret *messages.SignedMessage
	for _, m := range msgs {
		if ret == nil {
			ret = m.DeepCopy()
		} else {
			if err := ret.Aggregate(m); err != nil {
				return nil, errors.Wrap(err, "AggregateMsgs: could not aggregate msg")
			}
		}
	}
	return ret, nil
}

func isValidVCBCReady(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.VCBCReadyMsgType {
	// 	return errors.New("msg type is not VCBCReadyMsgType")
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

	VCBCReadyData, err := signedMsg.Message.GetVCBCReadyData()
	if err != nil {
		return errors.Wrap(err, "could not get VCBCReadyData data")
	}
	if err := VCBCReadyData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCReadyData invalid")
	}

	// author
	author := VCBCReadyData.Author
	authorInCommittee := false
	for _, opID := range operators {
		if opID.OperatorID == author {
			authorInCommittee = true
		}
	}
	if !authorInCommittee {
		return errors.New("author (OperatorID) doesn't exist in Committee")
	}

	// priority & hash
	// priority := VCBCReadyData.Priority
	// if state.VCBCState.HasM(author, priority) {
	// 	localHash, err := GetProposalsHash(state.VCBCState.GetM(author, priority))
	// 	if err != nil {
	// 		return errors.Wrap(err, "could not get local hash")
	// 	}
	// 	if !bytes.Equal(localHash, VCBCReadyData.Hash) {
	// 		return errors.New("existing (priority,author) proposals have different hash")
	// 	}
	// }

	return nil
}

func CreateVCBCReady(state *messages.State, config alea.IConfig, hash []byte, author types.OperatorID, signature []byte) (*messages.SignedMessage, error) {
	vcbcReadyData := &messages.VCBCReadyData{
		Hash:     hash,
		// Priority: priority,
		// Proof:			proof,
		Author: author,
		Signature: signature,
	}
	dataByts, err := vcbcReadyData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCReady: could not encode vcbcReadyData")
	}
	msg := &messages.Message{
		MsgType:    messages.VCBCReadyMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCReady: failed signing filler msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
