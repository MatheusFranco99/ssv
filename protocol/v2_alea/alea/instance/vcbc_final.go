package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/google/uuid"

	"bytes"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponVCBCFinal(signedMessage *messages.SignedMessage) error {

	// get Data
	vcbcFinalData, err := signedMessage.Message.GetVCBCFinalData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCFinal: could not get vcbcFinalData data from signedMessage")
	}

	// get sender ID
	senderID := signedMessage.GetSigners()[0]
	priority := vcbcFinalData.Priority
	author := vcbcFinalData.Author
	hash := vcbcFinalData.Hash
	aggregatedMsgs := vcbcFinalData.AggregatedMsg

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCFinal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(senderID)))
	}

	log("start")

	// if i.initTime == -1 || i.initTime == 0 {
	// 	i.initTime = makeTimestamp()
	// }

	log("check if has data")
	if !i.State.VCBCState.Has(author, priority) {
		return errors.New("Error: UponVCBCFinal: received vcbc final but don't have data to compare hashes.")
	}

	log("check hash")
	ownHash := i.State.VCBCState.GetHash(author, priority)
	if !bytes.Equal(ownHash, hash) {
		return errors.New("Error: UponVCBCFinal: hash different from local.")
	}

	log("check aggregated signature")
	// AggregatedMsg
	signedAggregatedMessage := &messages.SignedMessage{}
	signedAggregatedMessage.Decode(aggregatedMsgs)

	if err := signedAggregatedMessage.Signature.VerifyByOperators(signedAggregatedMessage, i.config.GetSignatureDomainType(), types.QBFTSignatureType, i.State.Share.Committee); err != nil {
		return errors.Wrap(err, "Error: UponVCBCFinal: aggregatedMsg signature invalid.")
	}
	if len(signedAggregatedMessage.GetSigners()) < int(i.State.Share.Quorum) {
		return errors.New("Error: UponVCBCFinal: aggregatedMsg signers length don't reach quorum.")
	}

	hasSentInit := i.State.ACState.HasSentInit(author, priority, specalea.FirstRound, byte(1))
	log(fmt.Sprintf("has sent init: %v", hasSentInit))
	if !hasSentInit {

		log("creating aba init")
		initMsg, err := CreateABAInit(i.State, i.config, byte(1), specalea.FirstRound, author, priority)
		if err != nil {
			return errors.Wrap(err, "uponABAConf: failed to create ABA Init message")
		}
		log("broadcast start abainit")
		i.Broadcast(initMsg)
		log("broadcast finish abainit")
	}
	log("finish")

	return nil
}

func isValidVCBCFinal(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.VCBCFinalMsgType {
	// 	return errors.New("msg type is not VCBCFinalMsgType")
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

	VCBCFinalData, err := signedMsg.Message.GetVCBCFinalData()
	if err != nil {
		return errors.Wrap(err, "could not get VCBCFinalData data")
	}
	if err := VCBCFinalData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCFinalData invalid")
	}

	// author
	author := VCBCFinalData.Author
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
	// priority := VCBCFinalData.Priority
	// if state.VCBCState.HasM(author, priority) {
	// 	localHash, err := GetProposalsHash(state.VCBCState.GetM(author, priority))
	// 	if err != nil {
	// 		return errors.Wrap(err, "could not get local hash")
	// 	}
	// 	if !bytes.Equal(localHash, VCBCFinalData.Hash) {
	// 		return errors.New("existing (priority,author) proposals have different hash")
	// 	}
	// }

	// AggregatedMsg
	// aggregatedMsg := VCBCFinalData.AggregatedMsg
	// signedAggregatedMessage := &messages.SignedMessage{}
	// signedAggregatedMessage.Decode(aggregatedMsg)

	// if err := signedAggregatedMessage.Signature.VerifyByOperators(signedAggregatedMessage, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
	// 	return errors.Wrap(err, "aggregatedMsg signature invalid")
	// }
	// if len(signedAggregatedMessage.GetSigners()) < int(state.Share.Quorum) {
	// 	return errors.New("aggregatedMsg signers don't reach quorum")
	// }

	// AggregatedMsg data
	// aggrMsgData := &specalea.VCBCReadyData{}
	// err = aggrMsgData.Decode(signedAggregatedMessage.Message.Data)
	// if err != nil {
	// 	return errors.Wrap(err, "could get VCBCReadyData from aggregated msg data")
	// }
	// if !bytes.Equal(VCBCFinalData.Hash, aggrMsgData.Hash) {
	// 	return errors.New("aggregated message has different hash than the given on the message")
	// }
	// if aggrMsgData.Author != VCBCFinalData.Author {
	// 	return errors.New("aggregated message has different author than the given on the message")
	// }
	// if aggrMsgData.Priority != VCBCFinalData.Priority {
	// 	return errors.New("aggregated message has different priority than the given on the message")
	// }

	return nil
}

func CreateVCBCFinal(state *messages.State, config alea.IConfig, hash []byte, priority specalea.Priority, aggregatedMsg []byte, author types.OperatorID) (*messages.SignedMessage, error) {
	vcbcFinalData := &specalea.VCBCFinalData{
		Hash:          hash,
		Priority:      priority,
		AggregatedMsg: aggregatedMsg,
		Author:        author,
	}
	dataByts, err := vcbcFinalData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode vcbcFinalData")
	}
	msg := &messages.Message{
		MsgType:    messages.VCBCFinalMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing filler msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
