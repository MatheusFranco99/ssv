package instance

import (
	"fmt"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/google/uuid"

	"math/rand"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponABAConf(signedABAConf *messages.SignedMessage) error {

	// get data
	ABAConfData, err := signedABAConf.Message.GetABAConfData()
	if err != nil {
		return errors.Wrap(err, "uponABAConf:could not get ABAConfData from signedABAConf")
	}

	// sender
	senderID := signedABAConf.GetSigners()[0]
	author := ABAConfData.Author
	priority := ABAConfData.Priority
	votes := ABAConfData.Votes
	round := ABAConfData.Round

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAConf "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("author", int(author)), zap.Int("priority", int(priority)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Binary("votes", votes))
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

	log("add conf")
	i.State.ACState.AddConf(author, priority, round, votes, senderID)

	log(fmt.Sprintf("len conf, quorum: %v %v", i.State.ACState.LenConf(author, priority, round), int(i.State.Share.Quorum)))
	if i.State.ACState.LenConf(author, priority, round) >= int(i.State.Share.Quorum) {

		log("will get coin")
		// coin := i.config.GetCoinF()(round)
		// coin := byte(1)
		coin := Coin(int(round), int(author), int(priority))
		log(fmt.Sprintf("coin: %v", coin))

		conf_values := i.State.ACState.GetConfValues(author, priority, round)
		log(fmt.Sprintf("conf values: %v", conf_values))

		init_vote := coin
		if len(conf_values) < 2 {
			init_vote = conf_values[0]
		}
		log(fmt.Sprintf("init vote : %v", init_vote))

		log(fmt.Sprintf("has sent init: %v", i.State.ACState.HasSentInit(author, priority, round+1, init_vote)))

		if !i.State.ACState.HasSentInit(author, priority, round+1, init_vote) {

			log("create aba init")
			initMsg, err := CreateABAInit(i.State, i.config, init_vote, round+1, author, priority)
			if err != nil {
				return errors.Wrap(err, "uponABAConf: failed to create ABA Init message")
			}
			log("broadcast start abainit")

			i.Broadcast(initMsg)

			log("broadcast finish abainit")

			log("set sent init and inc round ")
			i.State.ACState.SetSentInit(author, priority, round+1, init_vote)
			i.State.ACState.SetRound(author, priority, round+1)
		}

		log(fmt.Sprintf("len conf values == 1 and conf_values[0] == coin: %v %v", len(conf_values) == 1, conf_values[0] == coin))
		if len(conf_values) == 1 && conf_values[0] == coin {
			log(fmt.Sprintf("has sent finish: %v", i.State.ACState.HasSentFinish(author, priority, coin)))
			if !i.State.ACState.HasSentFinish(author, priority, coin) {

				log("create aba finish")
				finishMsg, err := CreateABAFinish(i.State, i.config, coin, author, priority)
				if err != nil {
					return errors.Wrap(err, "uponABAConf: failed to create ABA Finish message")
				}
				log("broadcast start abafinish")

				i.Broadcast(finishMsg)
				log("broadcast finish abafinish")

				log("set sent finish")
				i.State.ACState.SetSentFinish(author, priority, coin)

			}
		}

	}

	log("finish")

	return nil

	// // old message -> ignore
	// if ABAConfData.ACRound < i.State.ACState.ACRound {
	// 	return nil
	// }
	// if ABAConfData.Round < i.State.ACState.GetCurrentABAState().Round {
	// 	return nil
	// }
	// // if future round -> intialize future state
	// if ABAConfData.ACRound > i.State.ACState.ACRound {
	// 	i.State.ACState.InitializeRound(ABAConfData.ACRound)
	// }
	// if ABAConfData.Round > i.State.ACState.GetABAState(ABAConfData.ACRound).Round {
	// 	i.State.ACState.GetABAState(ABAConfData.ACRound).InitializeRound(ABAConfData.Round)
	// }

	// abaState := i.State.ACState.GetABAState(ABAConfData.ACRound)

	// // add the message to the containers
	// abaState.ABAConfContainer.AddMsg(signedABAConf)

	// alreadyReceived := abaState.HasConf(ABAConfData.Round, senderID)

	// // if never received this msg, update
	// if !alreadyReceived {
	// 	abaState.SetConf(ABAConfData.Round, senderID, ABAConfData.Votes)

	// }

	// // reached strong support -> try to decide value
	// if abaState.CountConf(ABAConfData.Round) >= i.State.Share.Quorum {

	// 	q := abaState.CountConfContainedInValues(ABAConfData.Round)
	// 	if q < i.State.Share.Quorum {
	// 		return nil
	// 	}

	// 	// get common coin
	// 	s := i.config.GetCoinF()(abaState.Round)

	// 	i.logger.Debug("$$$$$$ UponABAConf coin", zap.Int64("time(micro)", makeTimestamp()), zap.Binary("values", abaState.Values[ABAConfData.Round]), zap.Int("coin", int(s)), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Binary("votes", ABAConfData.Votes))

	// 	// if values = {0,1}, choose randomly (i.e. coin) value for next round
	// 	if len(abaState.Values[ABAConfData.Round]) == 2 {

	// 		abaState.SetVInput(ABAConfData.Round+1, s)

	// 	} else {
	// 		abaState.SetVInput(ABAConfData.Round+1, abaState.GetValues(ABAConfData.Round)[0])

	// 		// if value has only one value, sends FINISH
	// 		if abaState.GetValues(ABAConfData.Round)[0] == s {
	// 			// check if indeed never sent FINISH message for this vote
	// 			if !abaState.SentFinish(s) {
	// 				finishMsg, err := CreateABAFinish(i.State, i.config, s, ABAConfData.ACRound)
	// 				if err != nil {
	// 					return errors.Wrap(err, "uponABAConf: failed to create ABA Finish message")
	// 				}
	// 				i.logger.Debug("$$$$$$ UponABAConf broadcast start abafinish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(s)))

	// 				i.Broadcast(finishMsg)
	// 				i.logger.Debug("$$$$$$ UponABAConf broadcast finish abafinish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(s)))

	// 				// update sent finish flag
	// 				abaState.SetSentFinish(s, true)
	// 				// process own finish msg
	// 				// i.uponABAFinish(finishMsg)
	// 			}
	// 		}
	// 	}

	// 	// increment round
	// 	abaState.IncrementRound()

	// 	// start new round sending INIT message with vote
	// 	initMsg, err := CreateABAInit(i.State, i.config, abaState.GetVInput(abaState.Round), abaState.Round, ABAConfData.ACRound)
	// 	if err != nil {
	// 		return errors.Wrap(err, "uponABAConf: failed to create ABA Init message")
	// 	}
	// 	i.logger.Debug("$$$$$$ UponABAConf broadcast start abainit", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaState.GetVInput(abaState.Round))))

	// 	i.Broadcast(initMsg)

	// 	i.logger.Debug("$$$$$$ UponABAConf broadcast finish abainit", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Int("vote", int(abaState.GetVInput(abaState.Round))))

	// 	// update sent init flag
	// 	abaState.SetSentInit(abaState.Round, abaState.GetVInput(abaState.Round), true)
	// 	// process own aux msg
	// 	// i.uponABAInit(initMsg)
	// }

	// i.logger.Debug("$$$$$$ UponABAConf finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(ABAConfData.ACRound)), zap.Int("round", int(ABAConfData.Round)), zap.Int("sender", int(senderID)), zap.Binary("votes", ABAConfData.Votes))

	// return nil
}

func Coin(round int, author int, priority int) byte {
	// Set the seed
	rand.Seed(int64(round + author + priority))

	// Generate a random integer between 0 and 1
	result := byte(rand.Intn(2))

	return result
}

func isValidABAConf(
	state *messages.State,
	config alea.IConfig,
	signedMsg *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	// if signedMsg.Message.MsgType != specalea.ABAConfMsgType {
	// 	return errors.New("msg type is not ABAConfMsgType")
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

	ABAConfData, err := signedMsg.Message.GetABAConfData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAConfData data")
	}
	if err := ABAConfData.Validate(); err != nil {
		return errors.Wrap(err, "ABAConfData invalid")
	}

	// vote
	// votes := ABAConfData.Votes
	// for _, vote := range votes {
	// 	if vote != 0 && vote != 1 {
	// 		return errors.New("vote different than 0 and 1")
	// 	}
	// }

	return nil
}

func CreateABAConf(state *messages.State, config alea.IConfig, votes []byte, round specalea.Round, author types.OperatorID, priority specalea.Priority) (*messages.SignedMessage, error) {
	ABAConfData := &messages.ABAConfData{
		Votes:    votes,
		Round:    round,
		Author:   author,
		Priority: priority,
	}
	dataByts, err := ABAConfData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAConf: could not encode abaconf data")
	}
	msg := &messages.Message{
		MsgType:    messages.ABAConfMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAConf: failed signing abaconf msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
