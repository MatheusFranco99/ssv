package messages

import (
	// "fmt"

	// "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"bytes"
)

/*
 * CVState
 */



type CVVoteDataStruct struct {
	Data []byte
	DataAuthor types.OperatorID
	AggregatedSignature []byte
	NodeIDs []types.OperatorID
}

type CVState struct {
	round int
	MaxRound int
	values map[int]map[types.OperatorID]*CVVoteDataStruct
	Terminated bool
	EmptyDecision bool
	DecisionValue []byte
}

func NewCVVoteDataStruct(data []byte, dataAuthor types.OperatorID, aggSign []byte, nodeIDs []types.OperatorID) *CVVoteDataStruct {

	return &CVVoteDataStruct{
		Data: data,
		DataAuthor: dataAuthor,
		AggregatedSignature: aggSign,
		NodeIDs: nodeIDs,
	}
}


func NewCVState() *CVState {

	return &CVState{
		round: 0,
		MaxRound: 2,
		values: make(map[int]map[types.OperatorID]*CVVoteDataStruct),
		Terminated: false,
		EmptyDecision: false,
		DecisionValue: make([]byte,0),
	}
}

func (s *CVState) IsTerminated() bool {
	return s.Terminated
}
func (s *CVState) SetTerminated() {
	s.Terminated = true
}
func (s *CVState) SetEmptyDecision() {
	s.EmptyDecision = true
}
func (s *CVState) IsDecisionEmptyValue() bool {
	return s.EmptyDecision
}
func (s *CVState) SetDecisionValue(v []byte) {
	s.DecisionValue = v
}
func (s *CVState) GetDecisionValue() []byte {
	return s.DecisionValue
}

func (s *CVState) GetRound() int {
	return s.round
}
func (s *CVState) BumpRound() {
	s.round += 1
}
func (s *CVState) GetMaxRound() int {
	return s.MaxRound
}
func (s *CVState) Add(round int, sender types.OperatorID, value []byte, dataAuthor types.OperatorID, aggSign []byte, nodeIDs []types.OperatorID) {
	if _,ok := s.values[round]; !ok {
		s.values[round] = make(map[types.OperatorID]*CVVoteDataStruct)
	}
	s.values[round][sender] = NewCVVoteDataStruct(value,dataAuthor,aggSign,nodeIDs)
}
func (s *CVState) GetLen(round int) int {
	if _,ok := s.values[round]; !ok {
		s.values[round] = make(map[types.OperatorID]*CVVoteDataStruct)
	}
	return len(s.values[round])
}

func (s *CVState) GetVoteDataByData(round int, data []byte) *CVVoteDataStruct {

	if _,ok := s.values[round]; !ok {
		return nil
	}

	for _, vote_data := range s.values[round] {
		if bytes.Equal(data,vote_data.Data) {
			return vote_data
		}
	}
	return nil
}

func (s *CVState) GetVote(round int) ([]byte,int) {

	if _,ok := s.values[round]; !ok {
		return []byte{},0
	}

	countMap := make(map[string]int)
	for _, vote_data := range s.values[round] {
		str := string(vote_data.Data)
		countMap[str]++
	}
	
	maxCount := 0
	maxBytes := []byte{}

	for value, count := range countMap {
		if count > maxCount {
			maxCount = count
			maxBytes = []byte(value)
		}
	}
	return maxBytes,maxCount	
}
// func (s *CVState) () {
	
// }