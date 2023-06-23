package messages

import (
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

type ABASpecialState struct {
	Votes map[alea.ACRound]map[types.OperatorID]byte
	Quorum int
}

func NewABASpecialState(quorum int) *ABASpecialState {
	return &ABASpecialState{
		Votes: make(map[alea.ACRound]map[types.OperatorID]byte),
		Quorum: quorum,
	}
}

func (a *ABASpecialState) Add(round alea.ACRound, opID types.OperatorID, vote byte) {
	if _,ok := a.Votes[round]; !ok {
		a.Votes[round] = make(map[types.OperatorID]byte)
	}
	a.Votes[round][opID] = vote
}

func (a *ABASpecialState) HasQuorum(round alea.ACRound) bool {
	if _,ok := a.Votes[round]; !ok {
		return false
	}
	return len(a.Votes[round]) >= a.Quorum
}

func (a *ABASpecialState) IsSameVote(round alea.ACRound) bool {
	if _,ok := a.Votes[round]; !ok {
		return false
	}
	if len(a.Votes[round]) == 0 {
		return false
	}
	appeared_zero := false
	appeared_one := false
	for _, vote := range a.Votes[round] {
		if vote == byte(1) {
			appeared_one = true
		} else {
			appeared_zero = false
		}
		if (appeared_one && appeared_zero) {
			return false
		}
	}
	return !(appeared_one && appeared_zero)
}
