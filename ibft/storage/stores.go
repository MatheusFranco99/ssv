package storage

import (
	"sync"

	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	forksprotocol "github.com/MatheusFranco99/ssv/protocol/forks"
	aleastorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
	"github.com/MatheusFranco99/ssv/storage/basedb"
	"go.uber.org/zap"
)

func NewStoresFromRoles(db basedb.IDb, logger *zap.Logger, roles ...spectypes.BeaconRole) *ALEAStores {
	stores := NewStores()

	for _, role := range roles {
		stores.Add(role, New(db, logger, role.String(), forksprotocol.GenesisForkVersion))
	}

	return stores
}

// ALEAStores wraps sync map with cast functions to qbft store
type ALEAStores struct {
	m sync.Map
}

func NewStores() *ALEAStores {
	return &ALEAStores{
		m: sync.Map{},
	}
}

// Get store from sync map by role type
func (qs *ALEAStores) Get(role spectypes.BeaconRole) aleastorage.ALEAStore {
	s, ok := qs.m.Load(role)
	if ok {
		qbftStorage := s.(aleastorage.ALEAStore)
		return qbftStorage
	}
	return nil
}

// Add store to sync map by role as a key
func (qs *ALEAStores) Add(role spectypes.BeaconRole, store aleastorage.ALEAStore) {
	qs.m.Store(role, store)
}
