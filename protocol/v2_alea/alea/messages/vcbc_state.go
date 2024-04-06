package messages

import (
	// "fmt"

	// "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"bytes"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

/*
 * ReadyState
 */

type ReceivedReadys struct {
	// readys map[types.OperatorID][]byte
	// readys *PSigContainer
	readys    map[types.OperatorID]*SignedMessage
	sentFinal bool
	quorum    int
}

func NewReceivedReadys(quorum uint64) *ReceivedReadys {

	return &ReceivedReadys{
		readys:    make(map[types.OperatorID]*SignedMessage),
		sentFinal: false,
		quorum:    int(quorum),
	}
}

func (rs *ReceivedReadys) HasSentFinal() bool {
	return rs.sentFinal
}

func (rs *ReceivedReadys) SetSentFinal() {
	rs.sentFinal = true
}

func (rs *ReceivedReadys) Add(opID types.OperatorID, msg *SignedMessage) {
	if _, ok := rs.readys[opID]; !ok {
		rs.readys[opID] = msg
	}
}

func (rs *ReceivedReadys) HasQuorum() bool {
	return len(rs.readys) >= rs.quorum
}

func (rs *ReceivedReadys) GetLen() int {
	return len(rs.readys)
}

func (rs *ReceivedReadys) GetNodeIDs() []types.OperatorID {
	ans := make([]types.OperatorID, len(rs.readys))
	idx := 0
	for key, _ := range rs.readys {
		ans[idx] = key
		idx += 1
	}
	return ans
}

func (rs *ReceivedReadys) GetMessages() []*SignedMessage {
	ans := make([]*SignedMessage, len(rs.readys))
	idx := 0
	for _, m := range rs.readys {
		ans[idx] = m
		idx += 1
	}
	return ans
}

type SentReadys struct {
	data map[types.OperatorID][]byte
}

func NewSentReadys() *SentReadys {

	return &SentReadys{
		data: make(map[types.OperatorID][]byte),
	}
}

func (rs *SentReadys) Has(opID types.OperatorID) bool {
	_, ok := rs.data[opID]
	return ok
}

func (rs *SentReadys) Get(opID types.OperatorID) []byte {
	return rs.data[opID]
}

func (rs *SentReadys) EqualData(opID types.OperatorID, data []byte) bool {
	d, ok := rs.data[opID]
	if !ok {
		return false
	}

	if len(data) != len(d) {
		return false
	}
	return bytes.Equal(data, d)
}

func (rs *SentReadys) Add(opID types.OperatorID, data []byte) {
	if _, ok := rs.data[opID]; !ok {
		rs.data[opID] = data
	}
}

type VCBCData struct {
	Data              []byte
	Author            types.OperatorID
	Hash              []byte
	AggregatedMessage *SignedMessage
}

type VCBCState struct {
	numNodes int
	data     map[types.OperatorID]*VCBCData
}

func NewVCBCState(nodeIDs []types.OperatorID) *VCBCState {

	vcbcState := &VCBCState{
		numNodes: 0,
		data:     make(map[types.OperatorID]*VCBCData),
	}
	vcbcState.numNodes = len(nodeIDs)

	return vcbcState
}

func (v *VCBCState) ReInit(nodeIDs []types.OperatorID) {
	v.data = make(map[types.OperatorID]*VCBCData)
}

func (v *VCBCState) HasData(author types.OperatorID) bool {
	_, ok := v.data[author]
	return ok
}

func (v *VCBCState) GetDataFromAuthor(author types.OperatorID) []byte {
	if vcbcData, ok := v.data[author]; ok {
		return vcbcData.Data
	}
	return []byte{}
}

func (v *VCBCState) GetDataMap() map[types.OperatorID]*VCBCData {
	return v.data
}

func (v *VCBCState) SetVCBCData(author types.OperatorID, data []byte, hash []byte, aggregated_msg *SignedMessage) {
	if _, ok := v.data[author]; !ok {
		vcbcData := &VCBCData{
			Data:              data,
			Author:            author,
			Hash:              hash,
			AggregatedMessage: aggregated_msg,
		}
		v.data[author] = vcbcData
	}
}

func (v *VCBCState) GetVCBCData(author types.OperatorID) *VCBCData {
	if _, ok := v.data[author]; !ok {
		return nil
	}
	return v.data[author]
}

func (v *VCBCState) GetVCBCDataByData(data []byte) *VCBCData {

	for _, vcbcData := range v.data {
		if bytes.Equal(vcbcData.Data, data) {
			return vcbcData
		}
	}
	return nil
}
func (v *VCBCState) GetLen() int {
	return len(v.data)
}

func (v *VCBCState) AllEqual() bool {
	var ref_data []byte
	for _, vcbcData := range v.data {
		if ref_data == nil {
			ref_data = vcbcData.Data
		} else {
			if !bytes.Equal(vcbcData.Data, ref_data) {
				return false
			}
		}
	}
	return true
}

func (v *VCBCState) GetMaxValueOccurences() ([]byte, int) {
	countMap := make(map[string]int)

	for _, vcbcData := range v.data {
		str := string(vcbcData.Data)
		countMap[str]++
	}

	maxCount := 0
	maxBytes := []byte{}

	for data, count := range countMap {
		if count > maxCount {
			maxCount = count
			maxBytes = []byte(data)
		}
	}

	return maxBytes, maxCount
}

func (v *VCBCState) GetNodeIDs() []types.OperatorID {
	ans := make([]types.OperatorID, len(v.data))
	idx := 0
	for opID, _ := range v.data {
		ans[idx] = opID
		idx += 1
	}
	return ans
}
