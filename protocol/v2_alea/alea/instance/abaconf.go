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
	acround := ABAConfData.ACRound
	votes := ABAConfData.Votes
	round := ABAConfData.Round

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponABAConf "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(acround)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Binary("votes", votes))
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

	abaround.AddConf(votes, senderID)
	log("added conf")

	if i.State.ACState.CurrentACRound() < acround {
		log("future aba. quitting.")
		return nil
	}
	if aba.CurrentRound() < round {
		log("future aba round. quitting.")
		return nil
	}

	len_conf := abaround.LenConf()
	log(fmt.Sprintf("len conf: %v", len_conf))
	if len_conf >= int(i.State.Share.Quorum) {
		log("got conf quorum")

		
		// coin := i.config.GetCoinF()(round)
		// coin := byte(1)
		rand.Seed(int64(acround)+int64(round))
		coin := byte(rand.Intn(2))
		// coin := Coin(int(round), int(author), int(priority))

		// coin := i.State.CommonCoin.GetCoin(acround,round)
		log(fmt.Sprintf("coin: %v", coin))

		conf_values := abaround.GetConfValues()
		log(fmt.Sprintf("conf values: %v", conf_values))

		init_vote := coin
		if len(conf_values) < 2 {
			init_vote = conf_values[0]
		}
		log(fmt.Sprintf("init vote : %v", init_vote))

		has_sent_init := aba.GetABARound(round+1).HasSentInit(init_vote)
		log(fmt.Sprintf("has sent init: %v", has_sent_init))

		if !has_sent_init {

			initMsg, err := CreateABAInit(i.State, i.config, init_vote, round+1, acround)
			if err != nil {
				return errors.Wrap(err, "uponABAConf: failed to create ABA Init message")
			}
			log("created aba init")

			i.Broadcast(initMsg)
			log("broadcasted abainit")

			aba.GetABARound(round+1).SetSentInit(init_vote)
			aba.BumpRound()
			log("set sent init and inc round")
		}

		if len(conf_values) == 1 && conf_values[0] == coin {
			has_sent_finish := aba.HasSentFinish(coin)
			log(fmt.Sprintf("has sent finish: %v", has_sent_finish))
			if !has_sent_finish {

				finishMsg, err := CreateABAFinish(i.State, i.config, coin, acround)
				if err != nil {
					return errors.Wrap(err, "uponABAConf: failed to create ABA Finish message")
				}
				log("created aba finish")

				i.Broadcast(finishMsg)
				log("broadcasted abafinish")

				aba.SetSentFinish(coin)
				log("set sent finish")

			}
		}

	}

	log("finish")

	return nil
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
	if signedMsg.Message.MsgType != messages.ABAConfMsgType {
		return errors.New("msg type is not ABAConfMsgType")
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

	ABAConfData, err := signedMsg.Message.GetABAConfData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAConfData data")
	}
	if err := ABAConfData.Validate(); err != nil {
		return errors.Wrap(err, "ABAConfData invalid")
	}

	return nil
}

func CreateABAConf(state *messages.State, config alea.IConfig, votes []byte, round specalea.Round, acround specalea.ACRound) (*messages.SignedMessage, error) {
	ABAConfData := &messages.ABAConfData{
		Votes:    votes,
		Round:    round,
		ACRound: acround,
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
