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

func (i *Instance) uponABAAux(signedABAAux *messages.SignedMessage) error {

	// get message Data
	ABAAuxData, err := signedABAAux.Message.GetABAAuxData()
	if err != nil {
		return errors.Wrap(err, "uponABAAux: could not get ABAAuxData from signedABAAux")
	}

	// sender
	senderID := signedABAAux.GetSigners()[0]
	acround := ABAAuxData.ACRound
	vote := ABAAuxData.Vote
	round := ABAAuxData.Round

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAAux "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
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

	abaround.AddAux(vote, senderID)
	log("added aux")

	if i.State.ACState.CurrentACRound() < acround {
		log("future aba. quitting.")
		return nil
	}
	if aba.CurrentRound() < round {
		log("future aba round. quitting.")
		return nil
	}


	if abaround.HasSentConf() {
		log("already sent conf. quitting.")
		return nil
	}

	if abaround.LenAux() >= int(i.State.Share.Quorum) {
		log("got aux quorum.")

		confValues := abaround.GetConfValues()
		log(fmt.Sprintf("conf values: %v", confValues))

		confMsg, err := CreateABAConf(i.State, i.config, confValues, round, acround)
		if err != nil {
			return errors.Wrap(err, "uponABAAux: failed to create ABA Conf message after strong support")
		}
		log("created aba conf")

		i.Broadcast(confMsg)
		log("broadcasted")

		abaround.SetSentConf()
		log("set sent conf")
	}

	log("finish")
	return nil
}

func isValidABAAux(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != messages.ABAAuxMsgType {
		return errors.New("msg type is not ABAAuxMsgType")
	}
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	ABAAuxData, err := signedMsg.Message.GetABAAuxData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAAuxData data")
	}
	if err := ABAAuxData.Validate(); err != nil {
		return errors.Wrap(err, "ABAAuxData invalid")
	}

	return nil
}

func CreateABAAux(state *messages.State, config alea.IConfig, vote byte, round specalea.Round, acround specalea.ACRound) (*messages.SignedMessage, error) {
	ABAAuxData := &messages.ABAAuxData{
		Vote:     vote,
		Round:    round,
		ACRound: acround,
	}
	dataByts, err := ABAAuxData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAAux: could not encode abaaux data")
	}
	msg := &messages.Message{
		MsgType:    messages.ABAAuxMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAAux: failed signing abaaux msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
