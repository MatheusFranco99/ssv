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
	// data := vcbcFinalData.Data
	// author := vcbcFinalData.Author
	hash := vcbcFinalData.Hash
	aggregatedSignature := vcbcFinalData.AggregatedSignature
	nodeIDs := vcbcFinalData.NodesIds

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCFinal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)))
	}

	log("start")

	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	if !i.State.SentReadys.Has(senderID) {
		log("does not have data")
		return nil
	}

	data := i.State.SentReadys.Get(senderID)
	log("get data")

	i.State.VCBCState.SetVCBCData(senderID,data,hash,aggregatedSignature,nodeIDs)
	log("saved vcbc data")

	if (i.State.WaitForVCBCAfterDecided) {
		if (i.State.WaitForVCBCAfterDecided_Author == senderID) {
			log("it was waiting for such vcbc final to terminate")

			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data)
				log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			}
		}
	}



	if (i.State.VCBCState.GetLen() == int(len(i.State.Share.Committee))) {
		log("received N VCBC Final")
		if i.State.VCBCState.AllEqual() {
			log("all N VCBC are equal. Terminating.")
			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data)
				log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			}
		}
	}

	if (i.State.VCBCState.GetLen() >= int(i.State.Share.Quorum) && !i.State.StartedABA) {
		log("launching ABA")
		i.State.StartedABA = true
		err := i.StartABA()
		if err != nil {
			return err
		}
	}

	return nil
}

func isValidVCBCFinal(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
	logger *zap.Logger,
) error {
	// if signedMsg.Message.MsgType != specalea.VCBCFinalMsgType {
	// 	return errors.New("msg type is not VCBCFinalMsgType")
	// }

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		logger.Debug("$$$$$$ UponMV_VCBCFinal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

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

	VCBCFinalData, err := signedMsg.Message.GetVCBCFinalData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get VCBCFinalData data")
	}
	if err := VCBCFinalData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCFinalData invalid")
	}
	log("validated")

	// hash := VCBCFinalData.Hash
	// aggregatedSignature := VCBCFinalData.AggregatedSignature
	// nodeIDs := VCBCFinalData.NodesIds

	// pbkeys := make([][]byte,0)
	// for _, opID := range nodeIDs{
	// 	for _, operator := range operators {
	// 		if operator.OperatorID == opID {
	// 			pbkeys = append(pbkeys,operator.PubKey)
	// 			break
	// 		}
	// 	}
	// }

	// if (len(pbkeys) == 0) {
	// 	return errors.New("VCBCFinalData validation: didnt find any operator object with given operatorIDs")
	// }
	// if (len(pbkeys) != len(nodeIDs)) {
	// 	return errors.New("VCBCFinalData validation: list of pubkeys has different size than list of NodeIDs")
	// }

	// author := signedMsg.GetSigners()[0]
	// err = aggregatedSignature.VerifyMultiPubKey(messages.NewByteRoot(([]byte(fmt.Sprintf("%v%v",author,hash)))),config.GetSignatureDomainType(), types.QBFTSignatureType,pbkeys)
	// if err != nil {
	// 	return errors.Wrap(err,"VCBCFinalData validation: verification of aggregated signature failed.")
	// }

	return nil
}

func CreateVCBCFinal(state *messages.State, config alea.IConfig, hash []byte, aggSign []byte, nodeIDs []types.OperatorID) (*messages.SignedMessage, error) {
	
	vcbcFinalData := &messages.VCBCFinalData{
		Hash: hash,
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
