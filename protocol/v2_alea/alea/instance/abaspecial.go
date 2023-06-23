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

func (i *Instance) uponABASpecialVote(signedMsg *messages.SignedMessage) error {

	// get data
	abaSpecialVoteData, err := signedMsg.Message.GetABASpecialVoteData()
	if err != nil {
		return errors.Wrap(err, "uponABASpecialVote: could not get abaSpecialVoteData from signed message")
	}

	senderID := signedMsg.GetSigners()[0]
	acround := abaSpecialVoteData.ACRound
	vote := abaSpecialVoteData.Vote

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABASpecialVote "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(acround)), zap.Int("vote", int(vote)))
	}

	log("start")

	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}

	if i.State.ACState.CurrentACRound() > acround {
		log("old acround. quitting.")
		return nil
	}

	i.State.ABASpecialState.Add(acround, senderID, vote)
	log("added to abaspecialstate")

	if i.State.ACState.CurrentACRound() < acround {
		log("future acround. quitting.")
		return nil
	}

	if !(i.State.ABASpecialState.HasQuorum(acround)) {
		log("dont have quorum. quitting.")
		return nil
	}

	if !(i.State.ABASpecialState.IsSameVote(acround)) {
		log("have different votes. nothing to do.")
		return nil
	}


	aba := i.State.ACState.GetABA(acround)
	has_sent_finish := aba.HasSentFinish(vote)

	log(fmt.Sprintf("has sent finish: %v", has_sent_finish))
	if !has_sent_finish {

		finishMsg, err := CreateABAFinish(i.State, i.config, vote, acround)
		if err != nil {
			return errors.Wrap(err, "UponABASpecialVote: failed to create ABA Finish message")
		}
		log("created aba finish")

		i.Broadcast(finishMsg)
		log("broadcasted abafinish")

		aba.SetSentFinish(vote)
		log("set sent finish")

	}

	if vote == byte(0) {
		log("vote is 0. starting next ABA.")
		// i.StartABA()
		return nil
	}


	leader := i.State.Share.Committee[int(acround)%len(i.State.Share.Committee)].OperatorID
	log("calculted leader again")

	has_vcbc_final := i.State.VCBCState.HasData(leader)
	log(fmt.Sprintf("checked if has data: %v",has_vcbc_final))
	if (!has_vcbc_final) {
		i.State.WaitForVCBCAfterDecided = true
		i.State.WaitForVCBCAfterDecided_Author = leader
		log("set variables to wait for VCBC of leader of ABA decided with 1")
	} else {
		if !i.State.Decided {
			i.finalTime = makeTimestamp()
			diff := i.finalTime - i.initTime
			data := i.State.VCBCState.GetDataFromAuthor(leader)
			i.Decide(data)
			log(fmt.Sprintf("consensus decided. Total time: %v",diff))
		}
	}

	return nil
}

func isValidABASpecialVote(
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

	ABASpecialVoteData, err := signedMsg.Message.GetABASpecialVoteData()
	if err != nil {
		return errors.Wrap(err, "could not get ABASpecialVoteData data")
	}
	if err := ABASpecialVoteData.Validate(); err != nil {
		return errors.Wrap(err, "ABASpecialVoteData invalid")
	}

	return nil
}

func CreateABASpecialVote(state *messages.State, config alea.IConfig, vote byte, round specalea.ACRound) (*messages.SignedMessage, error) {
	ABASpecialVoteData := &messages.ABASpecialVoteData{
		Vote:     vote,
		ACRound:    round,
	}
	dataByts, err := ABASpecialVoteData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABASpecialVote: could not encode abainit data")
	}
	msg := &messages.Message{
		MsgType:    messages.ABASpecialVoteMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABASpecialVote: failed signing abainit msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
