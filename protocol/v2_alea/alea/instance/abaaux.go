package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponABAAux(signedABAAux *messages.SignedMessage) error {

	// Decode
	ABAAuxData, err := signedABAAux.Message.GetABAAuxData()
	if err != nil {
		return errors.Wrap(err, "uponABAAux: could not get ABAAuxData from signedABAAux")
	}

	// Sender
	senderID := signedABAAux.GetSigners()[0]
	acround := ABAAuxData.ACRound
	vote := ABAAuxData.Vote
	round := ABAAuxData.Round

	// Funciton identifier
	i.State.AbaAuxLogTag += 1

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$ UponABAAux "+fmt.Sprint(i.State.AbaAuxLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
	}

	log("start")

	if i.State.ACState.IsTerminated() {
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

		confMsg, err := i.CreateABAConf(confValues, round, acround)
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
	logger *zap.Logger,
) error {

	// logger
	log := func(str string) {

		if state.HideLogs || state.HideValidationLogs || state.DecidedLogOnly {
			return
		}
		logger.Debug("$$$$$$ UponMV_ABAAux : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if signedMsg.Message.MsgType != messages.ABAAuxMsgType {
		return errors.New("msg type is not ABAAuxMsgType")
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

	ABAAuxData, err := signedMsg.Message.GetABAAuxData()
	log("got data")

	if err != nil {
		return errors.Wrap(err, "could not get ABAAuxData data")
	}
	if err := ABAAuxData.Validate(); err != nil {
		return errors.Wrap(err, "ABAAuxData invalid")
	}
	log("validated")

	return nil
}

func (i *Instance) CreateABAAux(vote byte, round specalea.Round, acround specalea.ACRound) (*messages.SignedMessage, error) {

	state := i.State

	ABAAuxData := &messages.ABAAuxData{
		Vote:    vote,
		Round:   round,
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
	sig, hash_map, err := i.Sign(msg)
	if err != nil {
		panic(err)
	}

	signedMsg := &messages.SignedMessage{
		Signature:          sig,
		Signers:            []types.OperatorID{state.Share.OperatorID},
		Message:            msg,
		DiffieHellmanProof: hash_map,
	}
	return signedMsg, nil
}
