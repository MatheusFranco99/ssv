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

func (i *Instance) uponABAFinish(signedABAFinish *messages.SignedMessage) error { //(bool, []byte, error) {

	// get data
	ABAFinishData, err := signedABAFinish.Message.GetABAFinishData()
	if err != nil {
		return errors.Wrap(err, "uponABAFinish: could not get ABAFinishData from signedABAConf")
	}

	// sender
	senderID := signedABAFinish.GetSigners()[0]
	author := ABAFinishData.Author
	priority := ABAFinishData.Priority
	vote := ABAFinishData.Vote

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAFinish "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(senderID)), zap.Int("vote", int(vote)))
	}

	log("start")

	if i.initTime == -1 || i.initTime == 0 {
		i.initTime = makeTimestamp()
	}

	if i.State.ACState.CurrentPriority(author) > priority {
		log("finish old priority")
		return nil
	}

	log("add finish")
	i.State.ACState.AddFinish(author, priority, vote, senderID)

	log(fmt.Sprintf("len finish, partial quorum: %v %v", i.State.ACState.LenFinish(author, priority, vote), int(i.State.Share.PartialQuorum)))
	if i.State.ACState.LenFinish(author, priority, vote) >= int(i.State.Share.PartialQuorum) {

		log(fmt.Sprintf("has sent finish: %v", i.State.ACState.HasSentFinish(author, priority, vote)))
		if !i.State.ACState.HasSentFinish(author, priority, vote) {

			log("create aba finish")
			finishMsg, err := CreateABAFinish(i.State, i.config, vote, author, priority)
			if err != nil {
				return errors.Wrap(err, "uponABAFinish: failed to create ABA Finish message")
			}

			log("broadcast start abafinish")
			i.Broadcast(finishMsg)
			log("broadcast finish abafinish")

			log("set sent finish")
			i.State.ACState.SetSentFinish(author, priority, vote)

		}
	}

	log(fmt.Sprintf("len finish, quorum, is decided: %v %v %v", i.State.ACState.LenFinish(author, priority, vote), int(i.State.Share.Quorum), i.State.ACState.IsDecided(author, priority)))
	if i.State.ACState.LenFinish(author, priority, vote) >= int(i.State.Share.Quorum) && !i.State.ACState.IsDecided(author, priority) {

		// msgId := spectypesalea.MessageIDFromBytes(i.State.ID)
		// diff := i.timer.endTime(fmt.Sprintf("%s%s%v", hex.EncodeToString(msgId.GetPubKey()), msgId.GetRoleType().String(), int(i.State.Height)))
		i.finalTime = makeTimestamp()

		diff := i.finalTime - i.initTime
		// i.mu.Lock()
		// if val, ok := i.timeMap[fmt.Sprintf("%s%s%v", hex.EncodeToString(msgId.GetPubKey()), msgId.GetRoleType().String(), int(i.State.Height))]; ok {
		// 	// delete(timeMap, key) // remove the key from the map
		// 	diff = makeTimestamp() - val
		// }
		// i.mu.Unlock()

		log(fmt.Sprintf("set decided, set priority. Total time: %v", diff))
		i.State.ACState.SetDecided(author, priority, vote)
		i.State.ACState.SetPriority(author, priority+1)
		log("terminated aba")
		return nil

		// if author == i.State.Share.OperatorID {
		// i.logger.Debug("$$$$$$ UponABAFinish terminated aba", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)), zap.Int("vote", int(ABAFinishData.Vote)))
		// }
	}
	log("finish")

	return nil

	// // old message -> ignore
	// if ABAFinishData.ACRound < i.State.ACState.ACRound {
	// 	return nil
	// }
	// // if future round -> intialize future state
	// if ABAFinishData.ACRound > i.State.ACState.ACRound {
	// 	i.State.ACState.InitializeRound(ABAFinishData.ACRound)
	// }

	// abaState := i.State.ACState.GetABAState(ABAFinishData.ACRound)

	// // add the message to the container
	// abaState.ABAFinishContainer.AddMsg(signedABAFinish)

	// alreadyReceived := abaState.HasFinish(senderID)
	// // if never received this msg, update
	// if !alreadyReceived {

	// 	// get vote from FINISH message
	// 	vote := ABAFinishData.Vote

	// 	// increment counter
	// 	abaState.SetFinish(senderID, vote)
	// }

	// // if FINISH(b) reached partial quorum and never broadcasted FINISH(b), broadcast
	// if !abaState.SentFinish(byte(0)) && !abaState.SentFinish(byte(1)) {
	// 	for _, vote := range []byte{0, 1} {

	// 		if abaState.CountFinish(vote) >= i.State.Share.PartialQuorum {
	// 			// broadcast FINISH
	// 			finishMsg, err := CreateABAFinish(i.State, i.config, vote, ABAFinishData.ACRound)
	// 			if err != nil {
	// 				return errors.Wrap(err, "uponABAFinish: failed to create ABA Finish message")
	// 			}

	// 			i.logger.Debug("$$$$$$ UponABAFinish broadcast start", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAFinishData.ACRound)), zap.Int("sender", int(senderID)), zap.Int("vote", int(ABAFinishData.Vote)))

	// 			i.Broadcast(finishMsg)
	// 			i.logger.Debug("$$$$$$ UponABAFinish broadcast finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAFinishData.ACRound)), zap.Int("sender", int(senderID)), zap.Int("vote", int(ABAFinishData.Vote)))

	// 			// update sent flag
	// 			abaState.SetSentFinish(vote, true)
	// 			// process own finish msg
	// 			// i.uponABAFinish(finishMsg)
	// 		}
	// 	}
	// }

	// // if FINISH(b) reached Quorum, decide for b and send termination signal
	// for _, vote := range []byte{0, 1} {
	// 	if abaState.CountFinish(vote) >= i.State.Share.Quorum {
	// 		abaState.SetDecided(vote)
	// 		abaState.SetTerminate(true)

	// 		i.logger.Debug("$$$$$$ UponABAFinish terminated aba", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAFinishData.ACRound)), zap.Int("sender", int(senderID)), zap.Int("vote", int(ABAFinishData.Vote)))

	// 	}
	// }
	// i.logger.Debug("$$$$$$ UponABAFinish finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAFinishData.ACRound)), zap.Int("sender", int(senderID)), zap.Int("vote", int(ABAFinishData.Vote)))

	// return nil
}

func isValidABAFinish(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.ABAFinishMsgType {
	// 	return errors.New("msg type is not ABAFinishMsgType")
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

	ABAFinishData, err := signedMsg.Message.GetABAFinishData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAFinishData data")
	}
	if err := ABAFinishData.Validate(); err != nil {
		return errors.Wrap(err, "ABAFinishData invalid")
	}

	// vote
	// vote := ABAFinishData.Vote
	// if vote != 0 && vote != 1 {
	// 	return errors.New("vote different than 0 and 1")
	// }

	return nil
}

func CreateABAFinish(state *messages.State, config alea.IConfig, vote byte, author types.OperatorID, priority specalea.Priority) (*messages.SignedMessage, error) {
	ABAFinishData := &messages.ABAFinishData{
		Vote:     vote,
		Author:   author,
		Priority: priority,
	}
	dataByts, err := ABAFinishData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode abafinish data")
	}
	msg := &messages.Message{
		MsgType:    messages.ABAFinishMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing abafinish msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
