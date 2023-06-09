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

		i.State.ACState.SetDecided(author, priority, vote)
		i.State.ACState.SetPriority(author, priority+1)
		log("terminated aba")


		log(fmt.Sprintf("result: %v",int(vote)))
		
		if int(vote) == 1 {

			acround := int(i.State.ACState.ACRound)
			opIDList := make([]types.OperatorID, len(i.State.Share.Committee))
			for idx, op := range i.State.Share.Committee {
				opIDList[idx] = op.OperatorID
			}
			leader := opIDList[(acround)%len(opIDList)]

			has_vcbc_final := i.State.VCBCState.HasData(leader)
			if (!has_vcbc_final) {
				i.State.WaitForVCBCAfterDecided = true
				i.State.WaitForVCBCAfterDecided_Author = leader
			} else {
				i.finalTime = makeTimestamp()
				diff := i.finalTime - i.initTime
				log(fmt.Sprintf("consensus decided. Total time: %v",diff))
			}

		} else {
			i.State.ACState.IncrementRound()
			err := i.StartAlea()
			if err != nil {
				return err
			}
		}


		return nil
	}

	log("finish")

	return nil
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
