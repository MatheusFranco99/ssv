package instance

import (
	"bytes"

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
	senderID := signedMessage.GetSigners()[0]

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (i.State.DecidedLogOnly) {
			return
		}
		i.logger.Debug("$$$$$$ UponVCBCReady "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("sender", int(senderID)))
	}

	log("start")


	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}
	
	own_hash, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(i.StartValue)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: could not compute own hash")
	}
	log("computed own hash")

	if !bytes.Equal(own_hash,hash) {
		log("hash not equal. quitting.")
	}
	log("compared hashes")

	i.State.ReceivedReadys.Add(senderID,signature)
	log("added to received readys")



	has_sent_final := i.State.ReceivedReadys.HasSentFinal()
	if has_sent_final {
		log("has sent final. quitting.")
		return nil
	}

	len_readys := i.State.ReceivedReadys.GetLen()
	log(fmt.Sprintf("len readys: %v", len_readys))


	has_quorum := i.State.ReceivedReadys.HasQuorum()
	if !has_quorum {
		log("dont have quorum. quitting.")
		return nil
	}

	aggSignature, err := i.State.ReceivedReadys.AggregateSignatures(i.State.Share.ValidatorPubKey)
	if err != nil {
		return errors.Wrap(err,"uponVCBCReady: failed to reconstruct signature")
	}
	log("aggregated signatures")

	vcbcFinalMsg, err := CreateVCBCFinal(i.State, i.config, hash, aggSignature, i.State.ReceivedReadys.GetNodeIDs())
	if err != nil {
		return errors.Wrap(err, "uponVCBCReady: failed to create VCBCReady message with proof")
	}
	log("created vcbc final")

	i.Broadcast(vcbcFinalMsg)
	log("broadcasted")

	i.State.ReceivedReadys.SetSentFinal()
	log("set sent final.")


	return nil
}

func isValidVCBCReady(
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
		logger.Debug("$$$$$$ UponMV_VCBCReady "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	// if signedMsg.Message.MsgType != specalea.VCBCReadyMsgType {
	// 	return errors.New("msg type is not VCBCReadyMsgType")
	// }
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	log("checked height")
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	log("checked signers == 1")
	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}
	log("checked signature")

	VCBCReadyData, err := signedMsg.Message.GetVCBCReadyData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get VCBCReadyData data")
	}
	if err := VCBCReadyData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCReadyData invalid")
	}
	log("validated")

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
	log("checked author in committee")

	// hash := VCBCReadyData.Hash

	// signature := VCBCReadyData.Signature

	// opID := signedMsg.GetSigners()[0]
	// pbkey := []byte{}
	// for _, operator := range operators {
	// 	if operator.OperatorID == opID {
	// 		pbkey = operator.PubKey
	// 	}
	// }

	// if (len(pbkey) == 0) {
	// 	return errors.New("VCBCReadyData validation: didnt find operator object with given operatorID")
	// }

	// err = signature.Verify(messages.NewByteRoot([]byte(fmt.Sprintf("%v%v",author,hash))),config.GetSignatureDomainType(), types.QBFTSignatureType, pbkey)
	// if err != nil {
	// 	return errors.Wrap(err,"VCBCReadyData validation: verification of hash signature failed.")
	// }

	return nil
}

func CreateVCBCReady(state *messages.State, config alea.IConfig, hash []byte, author types.OperatorID, signature []byte) (*messages.SignedMessage, error) {
	vcbcReadyData := &messages.VCBCReadyData{
		Hash:     hash,
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
