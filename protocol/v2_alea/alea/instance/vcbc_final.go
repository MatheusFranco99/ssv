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
	"strings"
)

func (i *Instance) uponVCBCFinal(signedMessage *messages.SignedMessage) error {

	// get Data
	vcbcFinalData, err := signedMessage.Message.GetVCBCFinalData()
	if err != nil {
		return errors.Wrap(err, "uponVCBCFinal: could not get vcbcFinalData data from signedMessage")
	}

	// get sender ID
	senderID := signedMessage.GetSigners()[0]
	hash := vcbcFinalData.Hash
	aggregated_msg := vcbcFinalData.AggregatedMessage

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (i.State.DecidedLogOnly && !strings.Contains(str,"Total time")) {
			return
		}

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

	i.State.VCBCState.SetVCBCData(senderID,data,hash,aggregated_msg)//aggregatedSignature,nodeIDs)
	log("saved vcbc data")

	if (i.State.WaitForVCBCAfterDecided) {
		if (i.State.WaitForVCBCAfterDecided_Author == senderID) {
			log("it was waiting for such vcbc final to terminate")

			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data, signedMessage)
				log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			}
		}
	}



	if (i.State.EqualVCBCOptimization && i.State.VCBCState.GetLen() == int(len(i.State.Share.Committee))) {
		log("received N VCBC Final")
		if i.State.VCBCState.AllEqual() {
			log("all N VCBC are equal. Terminating.")
			if !i.State.Decided {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				i.Decide(data, signedMessage)
				log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			}
		}
	}

	if (!i.State.StartedABA) {
		if (!i.State.WaitVCBCQuorumOptimization || (i.State.VCBCState.GetLen() >= int(i.State.Share.Quorum)) ) {
			log("launching ABA")
			i.State.StartedABA = true
			err := i.StartABA()
			if err != nil {
				return err
			}
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

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (state.DecidedLogOnly) {
			return
		}
		logger.Debug("$$$$$$ UponMV_VCBCFinal "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")


	if signedMsg.Message.MsgType != messages.VCBCFinalMsgType {
		return errors.New("msg type is not VCBCFinalMsgType")
	}
	log("checked msg type")

	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	log("checked height")
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	log("checked signers == 1")
	// if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
	// 	return errors.Wrap(err, "msg signature invalid")
	// }
	// log("checked signature")


	msg_bytes, err := signedMsg.Message.Encode()
	if err != nil {
		return errors.Wrap(err,"Could not encode message.")
	}
	if !state.DiffieHellmanContainerOneTimeCost.VerifyHash(msg_bytes,signedMsg.GetSigners()[0],signedMsg.DiffieHellmanProof[state.Share.OperatorID]) {
		return errors.New("Failed Diffie Hellman verification")
	}


	VCBCFinalData, err := signedMsg.Message.GetVCBCFinalData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get VCBCFinalData data")
	}
	if err := VCBCFinalData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCFinalData invalid")
	}
	log("validated")

	aggregated_msg := VCBCFinalData.AggregatedMessage


	// verify signature
	if err := aggregated_msg.Signature.VerifyByOperators(aggregated_msg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "aggregated msg signature invalid")
	}
	log("checked aggregated message")

	return nil
}

func CreateVCBCFinal(state *messages.State, config alea.IConfig, hash []byte, aggregated_msg *messages.SignedMessage) (*messages.SignedMessage, error) {
	
	vcbcFinalData := &messages.VCBCFinalData{
		Hash: hash,
		AggregatedMessage: aggregated_msg,
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
	
	// No signing -> use Diffie Hellman
	// sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed signing filler msg")
	// }

	sig := []byte{}

	msg_bytes, err := msg.Encode()
	if err != nil {
		return nil, errors.Wrap(err,"CreateVCBCFinal: failed to encode message")
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
