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
	"strings"
)

func (i *Instance) uponABAFinish(signedABAFinish *messages.SignedMessage) error { //(bool, []byte, error) {
	
	// get data
	ABAFinishData, err := signedABAFinish.Message.GetABAFinishData()
	if err != nil {
		return errors.Wrap(err, "uponABAFinish: could not get ABAFinishData from signedABAConf")
	}

	// sender
	senderID := signedABAFinish.GetSigners()[0]
	acround := ABAFinishData.ACRound
	vote := ABAFinishData.Vote

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (i.State.DecidedLogOnly && !strings.Contains(str,"Total time")) {
			return
		}
		i.logger.Debug("$$$$$$ UponABAFinish "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("vote", int(vote)))
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

	aba := i.State.ACState.GetABA(acround)

	if aba.IsDecided() {
		log("aba already decided. quitting.")
		return nil
	}

	aba.AddFinish(vote, senderID)
	log("added finish")


	if i.State.ACState.CurrentACRound() < acround {
		log("future aba. quitting.")
		return nil
	}

	len_finish := aba.LenFinish(vote)
	log(fmt.Sprintf("len finish: %v",len_finish))

	if len_finish >= int(i.State.Share.PartialQuorum) {
		log("got finish partial quorum")

		has_sent_finish := aba.HasSentFinish(vote)
		log(fmt.Sprintf("has sent finish: %v", has_sent_finish))
		if !has_sent_finish {
			
			finishMsg, err := CreateABAFinish(i.State, i.config, vote, acround)
			if err != nil {
				return errors.Wrap(err, "uponABAFinish: failed to create ABA Finish message")
			}
			log("createed aba finish")

			i.Broadcast(finishMsg)
			log("broadcasted abafinish")

			aba.SetSentFinish(vote)
			log("set sent finish")

		}
	}

	if len_finish >= int(i.State.Share.Quorum) {
		log("got finish quorum")

		aba.SetDecided(vote)
		log(fmt.Sprintf("set aba to decided. Result: %v",int(vote)))
		
		if int(vote) == 1 {

			// acround := int(i.State.ACState.ACRound)
			// opIDList := make([]types.OperatorID, len(i.State.Share.Committee))
			// for idx, op := range i.State.Share.Committee {
			// 	opIDList[idx] = op.OperatorID
			// }
			// leader := opIDList[(acround)%len(opIDList)]
			leader := i.State.Share.Committee[int(acround)%len(i.State.Share.Committee)].OperatorID
			log("recalculated leader")

			has_vcbc_final := i.State.VCBCState.HasData(leader)
			log(fmt.Sprintf("has vcbc final of leader: %v",has_vcbc_final))

			i.State.ACState.TerminateAC()

			if (!has_vcbc_final) {
				i.State.WaitForVCBCAfterDecided = true
				i.State.WaitForVCBCAfterDecided_Author = leader
				log("set waiting for vcbc")
			} else {
				if !i.State.Decided {
					i.finalTime = makeTimestamp()
					diff := i.finalTime - i.initTime
					data := i.State.VCBCState.GetDataFromAuthor(leader)
					i.Decide(data, signedABAFinish)
					log(fmt.Sprintf("consensus decided. Total time: %v",diff))
				}
			}

		} else {
			i.State.ACState.BumpACRound()
			err := i.StartABA()
			if err != nil {
				return err
			}
		}
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
	logger *zap.Logger,
) error {

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {

		if (state.DecidedLogOnly) {
			return
		}
		logger.Debug("$$$$$$ UponMV_ABAFinish "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if signedMsg.Message.MsgType != messages.ABAFinishMsgType {
		return errors.New("msg type is not ABAFinishMsgType")
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

	ABAFinishData, err := signedMsg.Message.GetABAFinishData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get ABAFinishData data")
	}
	if err := ABAFinishData.Validate(); err != nil {
		return errors.Wrap(err, "ABAFinishData invalid")
	}
	log("validated")

	return nil
}

func CreateABAFinish(state *messages.State, config alea.IConfig, vote byte, acround specalea.ACRound) (*messages.SignedMessage, error) {
	ABAFinishData := &messages.ABAFinishData{
		Vote:     vote,
		ACRound: acround,
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
