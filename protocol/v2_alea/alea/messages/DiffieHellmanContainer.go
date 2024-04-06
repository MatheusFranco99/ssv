package messages

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"
)

type DiffieHellmanContainer struct {
	PublicShare    map[types.OperatorID]int
	CommonKey      map[types.OperatorID]int
	OwnPublicShare int
	a              int
	p              int
	g              int
}

func NewDiffieHellmanContainer() *DiffieHellmanContainer {
	return &DiffieHellmanContainer{
		PublicShare:    make(map[types.OperatorID]int),
		CommonKey:      make(map[types.OperatorID]int),
		OwnPublicShare: -1,
		a:              -1,
		p:              23,
		g:              5,
	}
}

func (d *DiffieHellmanContainer) GeneratePublicShare() {
	if d.OwnPublicShare != -1 {
		rand.Seed(time.Now().UnixNano())
		d.a = 1 + rand.Intn(d.p)
		d.OwnPublicShare = int(math.Pow(float64(d.g), float64(d.a))) % d.p
	}
}

func (d *DiffieHellmanContainer) GetPublicShare() int {
	return d.OwnPublicShare
}

func (d *DiffieHellmanContainer) HasPublicShare(opID types.OperatorID) bool {
	_, ok := d.PublicShare[opID]
	return ok
}

func (d *DiffieHellmanContainer) GetCommonKey(opID types.OperatorID) int {
	if v, ok := d.CommonKey[opID]; !ok {
		return v
	}
	return -1
}

func (d *DiffieHellmanContainer) AddPublicShare(opID types.OperatorID, value int) {
	if _, ok := d.PublicShare[opID]; !ok {
		d.PublicShare[opID] = value
		d.CommonKey[opID] = int(math.Pow(float64(value), float64(d.a))) % d.p
	}
}

func (d *DiffieHellmanContainer) GenerateHash(opID types.OperatorID, msg []byte) ([32]byte, error) {
	if _, ok := d.PublicShare[opID]; !ok {
		return [32]byte{}, errors.New("GenerateHash: Don't have opID public share.")
	}
	return sha256.Sum256([]byte(fmt.Sprintf("%v%v", msg, d.CommonKey[opID]))), nil
}
func (d *DiffieHellmanContainer) VerifyHash(opID types.OperatorID, hash []byte, msg []byte) (bool, error) {
	if _, ok := d.PublicShare[opID]; !ok {
		return false, errors.New("VerifyHash: Don't have opID public share.")
	}
	expected_hash := sha256.Sum256([]byte(fmt.Sprintf("%v%v", msg, d.CommonKey[opID])))
	return bytes.Equal(expected_hash[:], hash), nil
}