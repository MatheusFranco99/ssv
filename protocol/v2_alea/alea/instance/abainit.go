package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	"strings"

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
	i.State.AbaInitLogTag += 1

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.HideValidationLogs || (i.State.DecidedLogOnly && !strings.Contains(str, "Total time")) {
			return
		}
		i.logger.Debug("$$$$$$ UponABAInit "+fmt.Sprint(i.State.AbaInitLogTag)+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Int("vote", int(vote)))
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

	abaround.AddInit(vote, senderID)
	log("added init")

	if round == specalea.FirstRound {
		i.State.ABASpecialState.Add(acround, senderID, vote)
		log("added to abaspecialstate")
	}

	if i.State.ACState.CurrentACRound() < acround {
		log("future aba. quitting.")
		return nil
	}

	if i.State.FastABAOptimization && round == specalea.FirstRound {

		if i.State.ABASpecialState.HasQuorum(acround) {
			log("has quorum for special vote.")
			if i.State.ABASpecialState.IsSameVote(acround) {
				log("has equal votes.")

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
					i.StartABA()
				} else {
					leader := i.State.Share.Committee[int(acround)%len(i.State.Share.Committee)].OperatorID
					log("calculted leader again")

					has_vcbc_final := i.State.VCBCState.HasData(leader)
					log(fmt.Sprintf("checked if has data: %v", has_vcbc_final))
					if !has_vcbc_final {
						i.State.WaitForVCBCAfterDecided = true
						i.State.WaitForVCBCAfterDecided_Author = leader
						log("set variables to wait for VCBC of leader of ABA decided with 1")
					} else {
						if !i.State.Decided {
							i.finalTime = makeTimestamp()
							diff := i.finalTime - i.initTime
							data := i.State.VCBCState.GetDataFromAuthor(leader)
							i.Decide(data, signedABAInit)
							log(fmt.Sprintf("consensus decided. Total time: %v", diff))
							i.SendFinalForDecisionPurpose()
						}
					}
				}
			}
		}
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

		// Common coin for Conf quorum step
		if i.State.SendCommonCoin {
			i.SendCommonCoinShare()
		}

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

		abaround.SetSentInit(vote)
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

	// logger
	log := func(str string) {

		if state.HideLogs || state.HideValidationLogs || state.DecidedLogOnly {
			return
		}
		logger.Debug("$$$$$$ UponMV_ABAInit : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
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

	ABAInitData, err := signedMsg.Message.GetABAInitData()
	log("got data")
	if err != nil {
		return errors.Wrap(err, "could not get ABAInitData data")
	}
	if err := ABAInitData.Validate(); err != nil {
		return errors.Wrap(err, "ABAInitData invalid")
	}
	log("validated")

	// Signature will be checked outside
	// counter := state.AbaInitCounter
	// data := ABAInitData
	// if _, ok := counter[data.ACRound]; !ok {
	// 	counter[data.ACRound] = make(map[specalea.Round]uint64)
	// }
	// counter[data.ACRound][data.Round] += 1
	// if !(state.UseBLS) || !(state.AggregateVerify) || (counter[data.ACRound][data.Round] == state.Share.Quorum || counter[data.ACRound][data.Round] == state.Share.PartialQuorum) {
	// 	Verify(state, config, signedMsg, operators)
	// 	log("checked signature")
	// }

	return nil
}

func CreateABAInit(state *messages.State, config alea.IConfig, vote byte, round specalea.Round, acround specalea.ACRound) (*messages.SignedMessage, error) {
	ABAInitData := &messages.ABAInitData{
		Vote:    vote,
		Round:   round,
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
	sig, hash_map, err := Sign(state, config, msg)
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
