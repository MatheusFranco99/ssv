package instance

import (
	"fmt"

	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/google/uuid"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (i *Instance) uponCVVote(signedMsg *messages.SignedMessage) error {

	// get message Data
	msgData, err := signedMsg.Message.GetCVVoteData()
	if err != nil {
		return errors.Wrap(err, "uponCVVote: could not get CVData from signedMsg")
	}

	// sender
	senderID := signedMsg.GetSigners()[0]
	data := msgData.Data
	dataAuthor := msgData.DataAuthor
	aggSignature := msgData.AggregatedSignature
	nodeIDs := msgData.NodesIds
	round := msgData.Round

	//funciton identifier
	functionID := uuid.New().String()

	hash, err := GetDataHash(data)
	if err != nil {
		// log("error getting hash")
		return err
	}

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponCVVote "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("data_author", int(dataAuthor)), zap.Int("sender", int(senderID)), zap.Int("round", int(round)), zap.Binary("hash", hash))
	}

	log("start")

	if i.State.CVState.IsTerminated() {
		log("finish already terminated")
		return nil
	}

	if i.State.CVState.GetMaxRound() < round {
		log("finish round is bigger than max round")
		return nil
	}

	if i.State.CVState.GetRound() > round {
		log("finish old round")
		return nil
	}

	log("adding cv vote")
	i.State.CVState.Add(round,senderID,data, dataAuthor, aggSignature, nodeIDs)



	if i.State.CVState.GetRound() != round {
		log("finish future round")
		return nil
	}

	if i.State.CVState.GetLen(round) >= int(i.State.Share.Quorum) {
		
		new_vote, occurences := i.State.CVState.GetVote(round)
		new_vote_hash, err :=  GetDataHash(new_vote)
		if err != nil {
			log("error getting hash")
			return err
		}

		log(fmt.Sprintf("calculated own vote. Hash: %v, Occurences: %v", new_vote_hash, occurences))

		i.State.CVState.BumpRound()
		new_round := i.State.CVState.GetRound()

		if new_round <= i.State.CVState.GetMaxRound() {
			vcVoteData := i.State.CVState.GetVoteDataByData(round, new_vote)

			dataAuthor := vcVoteData.DataAuthor;
			aggSign := vcVoteData.AggregatedSignature;
			nodeIDs := vcVoteData.NodeIDs

			log(fmt.Sprintf("creating vc vote. Hash: %v",new_vote_hash))
			b_msg, err := CreateCVVote(i.State, i.config, new_vote, dataAuthor, aggSign, nodeIDs, new_round)
			if err != nil {
				return errors.Wrap(err, "UponCVVote: failed to create CV Vote message after strong support")
			}
			log("broadcast start")
			i.Broadcast(b_msg)
			log("broadcast finish")
		}


		if (occurences >= int(i.State.Share.Quorum)) {
			log("Occurences achieved strong support. Terminating consensus.")
			i.State.CVState.SetTerminated()
			i.State.CVState.SetDecisionValue(new_vote)

			i.finalTime = makeTimestamp()
			diff := i.finalTime - i.initTime
			log(fmt.Sprintf("consensus decided. Total time: %v",diff))
		} else {
			if new_round > i.State.CVState.GetMaxRound() {
				i.State.CVState.SetTerminated()
				i.State.CVState.SetEmptyDecision()
				log("Terminating CV with no consensus. Jumping to Alea.")
				err := i.StartAlea()
				if err != nil {
					return err
				}
			}
		}

	}


	return nil
}


func CreateCVVote(state *messages.State, config alea.IConfig, data []byte, author types.OperatorID, aggSignature []byte, nodeIDs []types.OperatorID, round int) (*messages.SignedMessage, error) {
	msgData := &messages.CVVoteData{
		Data: data,
		DataAuthor: author,
		AggregatedSignature: aggSignature,
		NodesIds: nodeIDs,
		Round: round,
	}
	dataByts, err := msgData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateCVVote: could not encode abaaux data")
	}
	msg := &messages.Message{
		MsgType:    messages.CVVoteMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateCVVote: failed signing abaaux msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
