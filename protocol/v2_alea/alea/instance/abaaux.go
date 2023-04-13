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
	author := ABAAuxData.Author
	priority := ABAAuxData.Priority
	vote := ABAAuxData.Vote
	round := ABAAuxData.Round

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAAux "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
	}

	log("start")

	if i.initTime == -1 || i.initTime == 0 {
		i.initTime = makeTimestamp()
	}

	if i.State.ACState.CurrentPriority(author) > priority {
		log("finish old priority")
		return nil
	}
	if i.State.ACState.CurrentRound(author, priority) > round {
		log("finish old round")
		return nil
	}

	log("add aux")
	i.State.ACState.AddAux(author, priority, round, vote, senderID)

	log(fmt.Sprintf("has sent conf: %v", i.State.ACState.HasSentConf(author, priority, round)))
	if i.State.ACState.HasSentConf(author, priority, round) {
		log("finish already sent conf")
		return nil
	}

	log(fmt.Sprintf("len aux, quorum: %v %v", i.State.ACState.LenAux(author, priority, round), int(i.State.Share.Quorum)))
	if i.State.ACState.LenAux(author, priority, round) >= int(i.State.Share.Quorum) {

		confValues := i.State.ACState.GetConfValues(author, priority, round)
		log(fmt.Sprintf("conf values: %v", confValues))
		// broadcast CONF message

		log("creating aba conf")
		confMsg, err := CreateABAConf(i.State, i.config, confValues, ABAAuxData.Round, author, priority)
		if err != nil {
			return errors.Wrap(err, "uponABAAux: failed to create ABA Conf message after strong support")
		}
		log("broadcast start")
		i.Broadcast(confMsg)
		log("broadcast finish")

		log("set sent conf")
		i.State.ACState.SetSentConf(author, priority, round)
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
	// if signedMsg.Message.MsgType != specalea.ABAAuxMsgType {
	// 	return errors.New("msg type is not ABAAuxMsgType")
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

	ABAAuxData, err := signedMsg.Message.GetABAAuxData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAAuxData data")
	}
	if err := ABAAuxData.Validate(); err != nil {
		return errors.Wrap(err, "ABAAuxData invalid")
	}

	// vote
	// vote := ABAAuxData.Vote
	// if vote != 0 && vote != 1 {
	// 	return errors.New("vote different than 0 and 1")
	// }

	return nil
}

func CreateABAAux(state *messages.State, config alea.IConfig, vote byte, round specalea.Round, author types.OperatorID, priority specalea.Priority) (*messages.SignedMessage, error) {
	ABAAuxData := &messages.ABAAuxData{
		Vote:     vote,
		Round:    round,
		Author:   author,
		Priority: priority,
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
