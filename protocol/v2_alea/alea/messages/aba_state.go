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
}

func NewABA() *ABA {
	aba := &ABA{
		abarounds:  make(map[alea.Round]*ABARound),
		round:      0,
		finish:     make(map[types.OperatorID]byte),
		sentFinish: make([]bool, 2),
	}
	aba.abarounds[alea.FirstRound] = NewABARound()
	return aba
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

func (ab *ABA) Len() int {
	return len(ab.abarounds)
}

func (ab *ABA) GetABARound(round alea.Round) *ABARound {
	return ab.abarounds[round]
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
	abas     map[types.OperatorID]map[alea.Priority]*ABA
	priority map[types.OperatorID]alea.Priority
	decided  map[types.OperatorID]map[alea.Priority]byte
	ACRound alea.ACRound
}

func NewACState(node_ids []types.OperatorID) *ACState {
	acstate := &ACState{
		abas:     make(map[types.OperatorID]map[alea.Priority]*ABA),
		priority: make(map[types.OperatorID]alea.Priority),
		decided:  make(map[types.OperatorID]map[alea.Priority]byte),
		ACRound: alea.FirstACRound,
	}
	for _, opID := range node_ids {
		acstate.abas[opID] = make(map[alea.Priority]*ABA)
		acstate.priority[opID] = alea.FirstPriority
		acstate.decided[opID] = make(map[alea.Priority]byte)
	}

	for _, opID := range node_ids {
		acstate.Init(opID, alea.FirstPriority)
	}

	return acstate
}

func (s *ACState) IncrementRound() {
	s.ACRound += 1
}

func (s *ACState) ReInit(node_ids []types.OperatorID) {
	s.abas = make(map[types.OperatorID]map[alea.Priority]*ABA)
	s.priority = make(map[types.OperatorID]alea.Priority)
	s.decided = make(map[types.OperatorID]map[alea.Priority]byte)

	for _, opID := range node_ids {
		s.abas[opID] = make(map[alea.Priority]*ABA)
		s.priority[opID] = alea.FirstPriority
		s.decided[opID] = make(map[alea.Priority]byte)
	}

	for _, opID := range node_ids {
		s.Init(opID, alea.FirstPriority)
	}
}

func (s *ACState) Init(id types.OperatorID, priority alea.Priority) {
	s.abas[id][priority] = NewABA()
}
func (s *ACState) CurrentPriority(author types.OperatorID) alea.Priority {
	return s.priority[author]
}
func (as *ACState) CurrentRound(author types.OperatorID, priority alea.Priority) alea.Round {
	if as.abas == nil {
		as.abas = make(map[types.OperatorID]map[alea.Priority]*ABA)
	}
	if _, ok := as.abas[author]; !ok {
		as.abas[author] = make(map[alea.Priority]*ABA)
	}
	if _, ok := as.abas[author][priority]; !ok {
		as.abas[author][priority] = NewABA()
	}
	return as.abas[author][priority].round
}

func (as *ACState) SetRound(author types.OperatorID, priority alea.Priority, round alea.Round) {
	if _, ok := as.abas[author][priority]; !ok {
		as.abas[author][priority] = NewABA()
	}
	as.abas[author][priority].round = round
}

func (as *ACState) SetDecided(author types.OperatorID, priority alea.Priority, vote byte) {
	if as.decided == nil {
		as.decided = make(map[types.OperatorID]map[alea.Priority]byte)
	}
	if _, ok := as.decided[author]; !ok {
		as.decided[author] = make(map[alea.Priority]byte)
	}
	if _, ok := as.decided[author][priority]; !ok {
		as.decided[author][priority] = vote
	}
}

func (as *ACState) IsDecided(author types.OperatorID, priority alea.Priority) bool {
	_, ok := as.decided[author][priority]
	return ok
}

func (as *ACState) SetPriority(author types.OperatorID, priority alea.Priority) {
	if as.priority == nil {
		as.priority = make(map[types.OperatorID]alea.Priority)
	}
	if _, ok := as.priority[author]; !ok {
		as.priority[author] = alea.FirstPriority
	}
	as.priority[author] = priority
}

func (as *ACState) SetupPriority(author types.OperatorID, priority alea.Priority) {
	if as.abas == nil {
		as.abas = make(map[types.OperatorID]map[alea.Priority]*ABA)
	}
	if _, ok := as.abas[author]; !ok {
		as.abas[author] = make(map[alea.Priority]*ABA)
	}
	if _, ok := as.abas[author][priority]; !ok {
		as.abas[author][priority] = NewABA()
	}
}

func (as *ACState) SetupRound(author types.OperatorID, priority alea.Priority, round alea.Round) {
	as.SetupPriority(author, priority)
	as.abas[author][priority].InitRound(round)
}

// --------------------------------------------------------------------------------

func (as *ACState) AddInit(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte, sender types.OperatorID) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).AddInit(vote, sender)
}

func (as *ACState) LenInit(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte) int {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).LenInit(vote)
}

func (as *ACState) SetSentInit(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).SetSentInit(vote)
}

func (as *ACState) HasSentInit(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte) bool {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).HasSentInit(vote)
}

// --------------------------------------------------------------------------------

func (as *ACState) AddAux(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte, sender types.OperatorID) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).AddAux(vote, sender)
}

func (as *ACState) LenAux(author types.OperatorID, priority alea.Priority, round alea.Round) int {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).LenAux()
}

func (as *ACState) SetSentAux(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).SetSentAux(vote)
}

func (as *ACState) HasSentAux(author types.OperatorID, priority alea.Priority, round alea.Round, vote byte) bool {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).HasSentAux(vote)
}

// --------------------------------------------------------------------------------

func (as *ACState) AddConf(author types.OperatorID, priority alea.Priority, round alea.Round, votes []byte, sender types.OperatorID) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).AddConf(votes, sender)
}

func (as *ACState) LenConf(author types.OperatorID, priority alea.Priority, round alea.Round) int {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).LenConf()
}

func (as *ACState) SetSentConf(author types.OperatorID, priority alea.Priority, round alea.Round) {
	as.SetupRound(author, priority, round)
	as.abas[author][priority].GetABARound(round).SetSentConf()
}

func (as *ACState) HasSentConf(author types.OperatorID, priority alea.Priority, round alea.Round) bool {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).HasSentConf()
}

func (as *ACState) GetConfValues(author types.OperatorID, priority alea.Priority, round alea.Round) []byte {
	as.SetupRound(author, priority, round)
	return as.abas[author][priority].GetABARound(round).GetConfValues()
}

// --------------------------------------------------------------------------------

func (as *ACState) AddFinish(author types.OperatorID, priority alea.Priority, vote byte, sender types.OperatorID) {
	as.abas[author][priority].AddFinish(vote, sender)
}

func (as *ACState) LenFinish(author types.OperatorID, priority alea.Priority, vote byte) int {
	return as.abas[author][priority].LenFinish(vote)
}

func (as *ACState) SetSentFinish(author types.OperatorID, priority alea.Priority, vote byte) {
	as.abas[author][priority].SetSentFinish(vote)
}

func (as *ACState) HasSentFinish(author types.OperatorID, priority alea.Priority, vote byte) bool {
	return as.abas[author][priority].HasSentFinish(vote)
}