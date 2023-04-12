package messages

import (
	"fmt"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

/*
 * ReadyState
 */

type ReadyState struct {
	readys map[alea.Priority]map[types.OperatorID]*SignedMessage
}

func NewReadyState() *ReadyState {

	readyState := &ReadyState{
		readys: make(map[alea.Priority]map[types.OperatorID]*SignedMessage),
	}
	return readyState
}

func (rs *ReadyState) ReInit() {
	rs.readys = make(map[alea.Priority]map[types.OperatorID]*SignedMessage)
}

func (rs *ReadyState) Add(priority alea.Priority, id types.OperatorID, msg *SignedMessage) {
	if rs.readys == nil {
		rs.readys = make(map[alea.Priority]map[types.OperatorID]*SignedMessage)
	}
	if _, ok := rs.readys[priority]; !ok {
		rs.readys[priority] = make(map[types.OperatorID]*SignedMessage)
	}
	rs.readys[priority][id] = msg
}

func (rs *ReadyState) GetLen(priority alea.Priority) int {
	if rs.readys == nil {
		rs.readys = make(map[alea.Priority]map[types.OperatorID]*SignedMessage)
	}
	if _, ok := rs.readys[priority]; !ok {
		return 0
	}
	return len(rs.readys[priority])
}

func (rs *ReadyState) GetMessages(priority alea.Priority) []*SignedMessage {
	if _, ok := rs.readys[priority]; !ok {
		rs.readys[priority] = make(map[types.OperatorID]*SignedMessage)
	}
	ans := make([]*SignedMessage, len(rs.readys[priority]))
	for i, msg := range rs.readys[priority] {
		ans[i] = (msg)
	}
	return ans
}

type VCBCState struct {
	numNodes int
	data     map[types.OperatorID]map[alea.Priority][]byte
}

func NewVCBCState(nodeIDs []types.OperatorID) *VCBCState {

	vcbcState := &VCBCState{
		numNodes: 0,
		data:     make(map[types.OperatorID]map[alea.Priority][]byte),
	}
	vcbcState.numNodes = len(nodeIDs)

	for _, id := range nodeIDs {
		vcbcState.data[id] = make(map[alea.Priority][]byte)
	}

	return vcbcState
}

func (v *VCBCState) ReInit(nodeIDs []types.OperatorID) {
	v.data = make(map[types.OperatorID]map[alea.Priority][]byte)
	for _, id := range nodeIDs {
		v.data[id] = make(map[alea.Priority][]byte)
	}
}

func (v *VCBCState) GetData() map[types.OperatorID]map[alea.Priority][]byte {
	return v.data
}

func (v *VCBCState) initAuthor(id types.OperatorID) {
	if v.data == nil {
		v.data = make(map[types.OperatorID]map[alea.Priority][]byte)
	}
	if _, ok := v.data[id]; !ok {
		v.data[id] = make(map[alea.Priority][]byte)
	}
}

func (v *VCBCState) Add(id types.OperatorID, priority alea.Priority, data []byte) {
	v.initAuthor(id)
	if _, ok := v.data[id][priority]; ok {
		fmt.Println("VCBCState: Trying to add to already existing priority")
	} else {
		v.data[id][priority] = data
	}
}

func (v *VCBCState) Has(id types.OperatorID, priority alea.Priority) bool {
	_, ok := v.data[id][priority]
	return ok
}

func (v *VCBCState) Get(id types.OperatorID, priority alea.Priority) []byte {
	if data, ok := v.data[id][priority]; ok {
		return data
	}
	return nil
}

func (v *VCBCState) Peek(id types.OperatorID) []byte {
	v.initAuthor(id)
	if len(v.data[id]) == 0 {
		return nil
	}

	minKey := -1
	for key := range v.data[id] {
		if minKey == -1 || int(key) < minKey {
			minKey = int(key)
		}
	}
	if minKey == -1 {
		return nil
	}
	return v.data[id][alea.Priority(minKey)]
}

func (v *VCBCState) Pop(id types.OperatorID) interface{} {
	if len(v.data[id]) == 0 {
		return nil
	}

	minKey := -1
	for key := range v.data[id] {
		if minKey == 0 || int(key) < minKey {
			minKey = int(key)
		}
	}
	if minKey == 0 {
		return nil
	}
	data := v.data[id][alea.Priority(minKey)]
	delete(v.data[id], alea.Priority(minKey))
	return data
}

func (v *VCBCState) Remove(id types.OperatorID, priority alea.Priority) {
	if _, ok := v.data[id][priority]; ok {
		delete(v.data[id], priority)
	}
}
