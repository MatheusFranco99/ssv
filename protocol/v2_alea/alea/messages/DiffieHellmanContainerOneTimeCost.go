package messages

import (
	// "bytes"
	"crypto/sha256"
	"fmt"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

type DiffieHellmanContainerOneTimeCost struct {
	own_operator_id int
	operators_ids []types.OperatorID
}

func NewDiffieHellmanContainerOneTimeCost(own_opID int,opIDs []types.OperatorID) *DiffieHellmanContainerOneTimeCost {
	return &DiffieHellmanContainerOneTimeCost{
		own_operator_id: own_opID,
		operators_ids: opIDs,
	}
}

func (d *DiffieHellmanContainerOneTimeCost) GetHashMap(value []byte) map[types.OperatorID][32]byte {
	ans := make(map[types.OperatorID][32]byte)
	for _, opID := range d.operators_ids {
		ans[opID] = sha256.Sum256([]byte(fmt.Sprintf("%v%v", value, int(opID) + int(d.own_operator_id))))
	}
	return ans
}

func (d *DiffieHellmanContainerOneTimeCost) VerifyHash(value []byte, sender types.OperatorID, hash [32]byte) bool {
	produced_hash := sha256.Sum256([]byte(fmt.Sprintf("%v%v", value, int(sender) + int(d.own_operator_id))))

	if len(hash) != len(produced_hash) {
		return false
	}
	for i,v := range hash {
		if v != produced_hash[i] {
			return false
		}
	}
	return true

	// return bytes.Equal(hash, sha256.Sum256([]byte(fmt.Sprintf("%v%v", value, int(sender) + int(d.own_operator_id)))))
}


