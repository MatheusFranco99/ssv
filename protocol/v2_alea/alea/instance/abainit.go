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
	author := abaInitData.Author
	priority := abaInitData.Priority
	round := abaInitData.Round
	vote := abaInitData.Vote

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAInit "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
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

	log("add init")
	i.State.ACState.AddInit(author, priority, round, vote, senderID)

	hasSentAux := i.State.ACState.HasSentAux(author, priority, round, vote)
	log(fmt.Sprintf("has sent aux: %v", hasSentAux))

	if hasSentAux {
		log("finish has already sent aux")
		return nil
	}

	log(fmt.Sprintf("len init, quorum: %v %v", i.State.ACState.LenInit(author, priority, round, vote), int(i.State.Share.Quorum)))

	if i.State.ACState.LenInit(author, priority, round, vote) >= int(i.State.Share.Quorum) {

		// sends AUX(b)
		log("creating aba aux")

		auxMsg, err := CreateABAAux(i.State, i.config, vote, round, author, priority)
		if err != nil {
			return errors.Wrap(err, "uponABAInit: failed to create ABA Aux message after strong init support")
		}
		log("set sent aux")
		i.State.ACState.SetSentAux(author, priority, round, vote)
		log("broadcast start aux")
		i.Broadcast(auxMsg)
		log("broadcast finish aux")
	}

	hasSentInit := i.State.ACState.HasSentInit(author, priority, round, vote)
	log(fmt.Sprintf("has sent init: %v", hasSentInit))

	if hasSentInit {
		log("finish has already sent this init")
		return nil
	}

	log(fmt.Sprintf("len init, partial quorum: %v %v", i.State.ACState.LenInit(author, priority, round, vote), int(i.State.Share.PartialQuorum)))
	if i.State.ACState.LenInit(author, priority, round, vote) >= int(i.State.Share.PartialQuorum) {

		// send INIT
		log("creating aba init")

		initMsg, err := CreateABAInit(i.State, i.config, vote, abaInitData.Round, author, priority)
		if err != nil {
			return errors.Wrap(err, "uponABAInit: failed to create ABA Init message after weak support")
		}
		log("set sent init")
		i.State.ACState.SetSentInit(author, priority, round, vote)

		log("broadcast start init")
		i.Broadcast(initMsg)
		log("broadcast finish init")
		// update sent flag
	}

	log("finish")

	// // old message -> ignore
	// if abaInitData.ACRound < i.State.ACState.ACRound {
	// 	return nil
	// }
	// if abaInitData.Round < i.State.ACState.GetCurrentABAState().Round {
	// 	return nil
	// }
	// // if future round -> intialize future state
	// if abaInitData.ACRound > i.State.ACState.ACRound {
	// 	i.State.ACState.InitializeRound(abaInitData.ACRound)
	// }
	// if abaInitData.Round > i.State.ACState.GetABAState(abaInitData.ACRound).Round {
	// 	i.State.ACState.GetABAState(abaInitData.ACRound).InitializeRound(abaInitData.Round)
	// }

	// abaState := i.State.ACState.GetABAState(abaInitData.ACRound)

	// // add the message to the container
	// abaState.ABAInitContainer.AddMsg(signedABAInit)

	// if !abaState.HasInit(abaInitData.Round, senderID, abaInitData.Vote) {
	// 	abaState.SetInit(abaInitData.Round, senderID, abaInitData.Vote)
	// }

	// // weak support -> send INIT
	// // if never sent INIT(b) but reached PartialQuorum (i.e. f+1, weak support), send INIT(b)
	// for _, vote := range []byte{0, 1} {
	// 	if !abaState.SentInit(abaInitData.Round, vote) && abaState.CountInit(abaInitData.Round, vote) >= i.State.Share.PartialQuorum {

	// 		// send INIT
	// 		initMsg, err := CreateABAInit(i.State, i.config, vote, abaInitData.Round, abaInitData.ACRound)
	// 		if err != nil {
	// 			return errors.Wrap(err, "uponABAInit: failed to create ABA Init message after weak support")
	// 		}

	// 		i.logger.Debug("$$$$$$ UponABAInit broadcast start init", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(abaInitData.ACRound)), zap.Int("round", int(abaInitData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaInitData.Vote)))
	// 		i.Broadcast(initMsg)
	// 		i.logger.Debug("$$$$$$ UponABAInit broadcast finish init", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(abaInitData.ACRound)), zap.Int("round", int(abaInitData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaInitData.Vote)))
	// 		// update sent flag
	// 		abaState.SetSentInit(abaInitData.Round, vote, true)
	// 		// process own init msg
	// 		// i.uponABAInit(initMsg)
	// 	}
	// }

	// // strong support -> send AUX
	// // if never sent AUX(b) but reached Quorum (i.e. 2f+1, strong support), sends AUX(b) and add b to values
	// for _, vote := range []byte{0, 1} {

	// 	if !abaState.SentAux(abaInitData.Round, vote) && abaState.CountInit(abaInitData.Round, vote) >= i.State.Share.Quorum {

	// 		// append vote

	// 		abaState.AddToValues(abaInitData.Round, vote)

	// 		// sends AUX(b)
	// 		auxMsg, err := CreateABAAux(i.State, i.config, vote, abaInitData.Round, abaInitData.ACRound)
	// 		if err != nil {
	// 			return errors.Wrap(err, "uponABAInit: failed to create ABA Aux message after strong init support")
	// 		}
	// 		i.logger.Debug("$$$$$$ UponABAInit broadcast start aux", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(abaInitData.ACRound)), zap.Int("round", int(abaInitData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaInitData.Vote)))
	// 		i.Broadcast(auxMsg)
	// 		i.logger.Debug("$$$$$$ UponABAInit broadcast finish aux", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(abaInitData.ACRound)), zap.Int("round", int(abaInitData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaInitData.Vote)))

	// 		// update sent flag
	// 		abaState.SetSentAux(abaInitData.Round, vote, true)
	// 		// process own aux msg
	// 		// i.uponABAAux(auxMsg)
	// 	}
	// }
	// i.logger.Debug("$$$$$$ UponABAInit finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(abaInitData.ACRound)), zap.Int("round", int(abaInitData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaInitData.Vote)))

	return nil
}

func isValidABAInit(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.ABAInitMsgType {
	// 	return errors.New("msg type is not specalea.ABAInitMsgType")
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

	ABAInitData, err := signedMsg.Message.GetABAInitData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAInitData data")
	}
	if err := ABAInitData.Validate(); err != nil {
		return errors.Wrap(err, "ABAInitData invalid")
	}

	// vote
	vote := ABAInitData.Vote
	if vote != 0 && vote != 1 {
		return errors.New("vote different than 0 and 1")
	}

	return nil
}

func CreateABAInit(state *messages.State, config alea.IConfig, vote byte, round specalea.Round, author types.OperatorID, priority specalea.Priority) (*messages.SignedMessage, error) {
	ABAInitData := &messages.ABAInitData{
		Vote:     vote,
		Round:    round,
		Author:   author,
		Priority: priority,
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
