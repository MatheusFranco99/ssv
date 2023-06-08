package messages

import (
	// "fmt"

	// "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"bytes"
)

/*
 * ReadyState
 */

type ReceivedReadys struct {
	readys map[types.OperatorID][]byte
	sentFinal bool
}

func NewReceivedReadys() *ReceivedReadys {

	return &ReceivedReadys{
		readys: make(map[types.OperatorID][]byte),
		sentFinal: false,
	}
}

func (rs *ReceivedReadys) HasSentFinal() bool {
	return rs.sentFinal
}

func (rs *ReceivedReadys) SetSentFinal() {
	rs.sentFinal = true;
}

func (rs *ReceivedReadys) Add(opID types.OperatorID, signature []byte) {
	rs.readys[opID] = signature
}

func (rs *ReceivedReadys) GetLen() int {
	return len(rs.readys)
}

func (rs *ReceivedReadys) GetSignatureMap() map[types.OperatorID][]byte {
	return rs.readys
}

func (rs *ReceivedReadys) GetNodeIDs() []types.OperatorID {
	ans := make([]types.OperatorID,len(rs.readys))
	idx := 0
	for key, _ := range rs.readys {
		ans[idx] = key
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

func (rs *SentReadys) EqualData(opID types.OperatorID, data []byte) bool {
	d, ok := rs.data[opID]
	if !ok {
		return false;
	}

	if len(data) != len(d) {
		return false;
	}
	return bytes.Equal(data,d)
}

func (rs *SentReadys) Add(opID types.OperatorID, data []byte) {
	if _, ok := rs.data[opID]; !ok {
		rs.data[opID] = data
	}
}



type VCBCData struct {
	Data []byte
	Author types.OperatorID
	Hash []byte
	AggregatedSignature []byte
	NodeIDs []types.OperatorID
}

type VCBCState struct {
	numNodes int
	data	 map[types.OperatorID]*VCBCData
}



func NewVCBCState(nodeIDs []types.OperatorID) *VCBCState {

	vcbcState := &VCBCState{
		numNodes: 0,
		data: make(map[types.OperatorID]*VCBCData),
	}
	vcbcState.numNodes = len(nodeIDs)

	// for _, id := range nodeIDs {
	// 	vcbcState.data[id] = make(map[alea.Priority][]byte)
	// 	vcbcState.hash[id] = make(map[alea.Priority][]byte)
	// }

	return vcbcState
}

func (v *VCBCState) ReInit(nodeIDs []types.OperatorID) {
	v.data = make(map[types.OperatorID]*VCBCData)
	// v.data = make(map[types.OperatorID]map[alea.Priority][]byte)
	// v.hash = make(map[types.OperatorID]map[alea.Priority][]byte)
	// for _, id := range nodeIDs {
	// 	v.data[id] = make(map[alea.Priority][]byte)
	// 	v.hash[id] = make(map[alea.Priority][]byte)
	// }
}

func (v *VCBCState) GetDataMap() map[types.OperatorID]*VCBCData {
	return v.data
}

func (v *VCBCState) SetVCBCData(author types.OperatorID, data []byte, hash []byte, aggregatedSignature []byte, nodeIDs []types.OperatorID) {
	if _, ok := v.data[author]; !ok {
		vcbcData := &VCBCData{
			Data: data,
			Author: author,
			Hash: hash,
			AggregatedSignature: aggregatedSignature,
			NodeIDs: nodeIDs,
		}
		v.data[author] = vcbcData
	}
}

func (v *VCBCState) GetVCBCData(author types.OperatorID) *VCBCData {
	if _, ok := v.data[author]; !ok {
		return nil;
	}
	return v.data[author]
}

func (v *VCBCState) GetVCBCDataByData(data []byte) *VCBCData {
	// if _, ok := v.data[author]; !ok {
	// 	return nil;
	// }
	// return v.data[author]

	for _, vcbcData := range v.data {
		if bytes.Equal(vcbcData.Data,data) {
			return vcbcData
		}
	}
	return nil
}
func (v *VCBCState) GetLen() int {
	return len(v.data)
}

func (v *VCBCState) GetMaxValueOccurences() ([]byte,int) {
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
