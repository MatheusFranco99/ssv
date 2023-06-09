package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/google/uuid"

	// "bytes"

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
	data := vcbcFinalData.Data
	author := vcbcFinalData.Author
	hash := vcbcFinalData.Hash
	aggregatedSignature := vcbcFinalData.AggregatedSignature
	nodeIDs := vcbcFinalData.NodesIds

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCFinal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("sender", int(senderID)))
	}

	log("start")

	log("set vcbc data")
	i.State.VCBCState.SetVCBCData(author,data,hash,aggregatedSignature,nodeIDs)

	log("check conditions to start cv")
	if (i.State.VCBCState.GetLen() >= int(i.State.Share.Quorum) && !i.State.StartedCV) {

		log("starting cv")
		i.State.StartedCV = true
		err := i.StartCV()
		if err != nil {
			return err
		}

	}

	log("check if it's waiting for such vcbc final to terminate")
	if (i.State.WaitForVCBCAfterDecided) {
		if (i.State.WaitForVCBCAfterDecided_Author == author) {

			i.finalTime = makeTimestamp()
			diff := i.finalTime - i.initTime
			log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			return nil
		}
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

func CreateVCBCFinal(state *messages.State, config alea.IConfig, data []byte, hash []byte, author types.OperatorID, aggSign []byte, nodeIDs []types.OperatorID) (*messages.SignedMessage, error) {
	
	vcbcFinalData := &messages.VCBCFinalData{
		Data: data,
		Hash: hash,
		Author: author,
		AggregatedSignature: aggSign,
		NodesIds: nodeIDs,
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
