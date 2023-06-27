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

func (i *Instance) uponABAInit(signedABAInit *messages.SignedMessage) error {

	
	// get data
	abaInitData, err := signedABAInit.Message.GetABAInitData()
	if err != nil {
		return errors.Wrap(err, "uponABAInit: could not get abainitdata from signedABAInit")
	}

	// sender
	senderID := signedABAInit.GetSigners()[0]
	acround := abaInitData.ACRound
	round := abaInitData.Round
	vote := abaInitData.Vote

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAInit "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
	}


	log("start")

	if (i.State.ACState.IsTerminated()) {
		log("ac terminated. quitting.")
		return nil
	}

	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}



	if i.State.ACState.CurrentACRound() > acround {
		log("old acround. quitting.")
		return nil
	}

	if i.State.ACState.GetABA(acround).CurrentRound() > round {
		log("old aba round. quitting.")
		return nil
	}


	aba := i.State.ACState.GetABA(acround)
	abaround := aba.GetABARound(round)

	abaround.AddInit(vote, senderID)
	log("added init")

	if i.State.ACState.CurrentACRound() < acround {
		log("future aba. quitting.")
		return nil
	}
	if aba.CurrentRound() < round {
		log("future aba round. quitting.")
		return nil
	}


	hasSentAux := abaround.HasSentAux(vote)
	log(fmt.Sprintf("has sent aux: %v", hasSentAux))

	if hasSentAux {
		log("finish has already sent aux for that vote. quitting")
		return nil
	}

	if abaround.LenInit(vote) >= int(i.State.Share.Quorum) {
		log("got init quorum.")

		auxMsg, err := CreateABAAux(i.State, i.config, vote, round, acround)
		if err != nil {
			return errors.Wrap(err, "uponABAInit: failed to create ABA Aux message after strong init support")
		}
		log("created aba aux")

		i.Broadcast(auxMsg)
		log("broadcasted aux")

		abaround.SetSentAux(vote)
		log("set sent aux")
	}

	hasSentInit := abaround.HasSentInit(vote)
	log(fmt.Sprintf("has sent init: %v", hasSentInit))

	if hasSentInit {
		log("has already sent this init. quitting.")
		return nil
	}

	if abaround.LenInit(vote) >= int(i.State.Share.PartialQuorum) {
		log("got init partial quorum.")

		// send INIT
		initMsg, err := CreateABAInit(i.State, i.config, vote, round, acround)
		if err != nil {
			return errors.Wrap(err, "uponABAInit: failed to create ABA Init message after weak support")
		}
		log("created aba init")

		i.Broadcast(initMsg)
		log("broadcasted init")

		abaround.SetSentInit( vote)
		log("set sent init")
	}

	log("finish")

	return nil
}

func isValidABAInit(
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
		logger.Debug("$$$$$$ UponMV_ABAInit "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if signedMsg.Message.MsgType != messages.ABAInitMsgType {
		return errors.New("msg type is not specalea.ABAInitMsgType")
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
	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}
	log("checked signature")

	ABAInitData, err := signedMsg.Message.GetABAInitData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get ABAInitData data")
	}
	if err := ABAInitData.Validate(); err != nil {
		return errors.Wrap(err, "ABAInitData invalid")
	}
	log("validated")

	return nil
}

func CreateABAInit(state *messages.State, config alea.IConfig, vote byte, round specalea.Round, acround specalea.ACRound) (*messages.SignedMessage, error) {
	ABAInitData := &messages.ABAInitData{
		Vote:     vote,
		Round:    round,
		ACRound: acround,
	}
	dataByts, err := ABAInitData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAInit: could not encode abainit data")
	}
	msg := &messages.Message{
		MsgType:    messages.ABAInitMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAInit: failed signing abainit msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
