package messages

import (
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

/*
 * ABARound
 */

type ABARound struct {
	init     map[types.OperatorID]byte
	aux      map[types.OperatorID]byte
	conf     map[types.OperatorID][]byte
	sentInit []bool
	sentAux  []bool
	sentConf bool
}

func NewABARound() *ABARound {
	abaround := &ABARound{
		init:     make(map[types.OperatorID]byte),
		aux:      make(map[types.OperatorID]byte),
		conf:     make(map[types.OperatorID][]byte),
		sentInit: make([]bool, 2),
		sentAux:  make([]bool, 2),
		sentConf: false,
	}
	return abaround
}

func (r *ABARound) AddInit(vote byte, sender types.OperatorID) {
	if _, ok := r.init[sender]; ok {
		return
	}
	r.init[sender] = vote
}

func (r *ABARound) LenInit(vote byte) int {
	ans := 0
	for _, v := range r.init {
		if v == vote {
			ans++
		}
	}
	return ans
}

func (r *ABARound) SetSentInit(vote byte) {
	if vote == 0 || vote == 1 {
		r.sentInit[vote] = true
	}
}

func (r *ABARound) HasSentInit(vote byte) bool {
	if vote == 0 || vote == 1 {
		return r.sentInit[vote]
	}
	return false
}

func (r *ABARound) AddAux(vote byte, sender types.OperatorID) {
	if _, ok := r.aux[sender]; ok {
		return
	}
	r.aux[sender] = vote
}

func (r *ABARound) LenAux() int {
	return len(r.aux)
}

func (r *ABARound) SetSentAux(vote byte) {
	if vote == 0 || vote == 1 {
		r.sentAux[vote] = true
	}
}

func (r *ABARound) HasSentAux(vote byte) bool {
	if vote == 0 || vote == 1 {
		return r.sentAux[vote]
	}
	return false
}

func (r *ABARound) AddConf(votes []byte, sender types.OperatorID) {
	if _, ok := r.conf[sender]; ok {
		return
	}
	r.conf[sender] = votes
}

func (r *ABARound) LenConf() int {
	return len(r.conf)
}

func (r *ABARound) SetSentConf() {
	r.sentConf = true
}

func (r *ABARound) HasSentConf() bool {
	return r.sentConf
}

func (r *ABARound) GetConfValues() []byte {
	values := make(map[byte]bool)
	for _, v := range r.aux {
		values[v] = true
	}
	result := make([]byte, 0, len(values))
	for k := range values {
		result = append(result, k)
	}
	return result
}

/*
 * ABA
 */

type ABA struct {
	abarounds  map[alea.Round]*ABARound
	round      alea.Round
	finish     map[types.OperatorID]byte
	sentFinish []bool
	decided		bool
	decidedValue	byte
}

func NewABA() *ABA {
	aba := &ABA{
		abarounds:  make(map[alea.Round]*ABARound),
		round:      alea.FirstRound,
		finish:     make(map[types.OperatorID]byte),
		sentFinish: make([]bool, 2),
		decided: false,
		decidedValue: byte(2),
	}
	aba.abarounds[alea.FirstRound] = NewABARound()
	return aba
}



func (ab *ABA) IsDecided() bool {
	return ab.decided
}
func (ab *ABA) SetDecided(vote byte) {
	ab.decided = true
	ab.decidedValue = vote
}
func (ab *ABA) GetDecidedValue() byte {
	return ab.decidedValue
}



func (ab *ABA) BumpRound() {
	ab.round += 1
	ab.InitRound(ab.round)
}




func (ab *ABA) InitRound(round alea.Round) {
	if _, ok := ab.abarounds[round]; !ok {
		ab.abarounds[round] = NewABARound()
	}
}

func (ab *ABA) AddRound(round alea.Round) {
	if _, ok := ab.abarounds[round]; !ok {
		ab.abarounds[round] = NewABARound()
	}
}

func (ab *ABA) GetABARound(round alea.Round) *ABARound {
	if _, ok := ab.abarounds[round]; !ok {
		ab.abarounds[round] = NewABARound()
	}
	return ab.abarounds[round]
}

func (ab *ABA) CurrentRound() alea.Round {
	return ab.round
}

func (ab *ABA) Len() int {
	return len(ab.abarounds)
}

func (ab *ABA) AddFinish(vote byte, sender types.OperatorID) {
	if _, ok := ab.finish[sender]; ok {
		return
	}
	ab.finish[sender] = vote
}

func (ab *ABA) LenFinish(vote byte) int {
	ans := 0
	for _, v := range ab.finish {
		if v == vote {
			ans++
		}
	}
	return ans
}

func (ab *ABA) SetSentFinish(vote byte) {
	if vote == 0 || vote == 1 {
		ab.sentFinish[vote] = true
	}
}

func (ab *ABA) HasSentFinish(vote byte) bool {
	if vote == 0 || vote == 1 {
		return ab.sentFinish[vote]
	}
	return false
}

/*
 * ACState
 */

type ACState struct {
	ABA		map[alea.ACRound]*ABA
	ACRound alea.ACRound
	Terminated bool
}

func NewACState() *ACState {
	acstate := &ACState{
		ABA: make(map[alea.ACRound]*ABA),
		ACRound: alea.FirstACRound,
		Terminated: false,
	}
	return acstate
}



func (s *ACState) IsTerminated() bool {
	return s.Terminated
}

func (s *ACState) TerminateAC() {
	s.Terminated = true
}

func (s *ACState) BumpACRound() {
	s.ACRound += 1
}

func (s *ACState) CurrentACRound() alea.ACRound {
	return s.ACRound
}

func (s *ACState) GetABA(acround alea.ACRound) *ABA {
	if _,ok := s.ABA[acround]; !ok {
		s.ABA[acround] = NewABA()
	}
	return s.ABA[acround]
}