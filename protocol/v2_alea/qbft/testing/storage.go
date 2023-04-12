package testing

import (
	"context"
	"sync"

	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"go.uber.org/zap"

	aleastorage "github.com/MatheusFranco99/ssv/ibft/storage"
	"github.com/MatheusFranco99/ssv/storage"
	"github.com/MatheusFranco99/ssv/storage/basedb"
)

var db basedb.IDb
var dbOnce sync.Once

func getDB() basedb.IDb {
	dbOnce.Do(func() {
		logger := zap.L()
		dbInstance, err := storage.GetStorageFactory(basedb.Options{
			Type:      "badger-memory",
			Path:      "",
			Reporting: false,
			Logger:    logger,
			Ctx:       context.TODO(),
		})
		if err != nil {
			panic(err)
		}
		db = dbInstance
	})
	return db
}

var allRoles = []spectypes.BeaconRole{
	spectypes.BNRoleAttester,
	spectypes.BNRoleProposer,
	spectypes.BNRoleAggregator,
	spectypes.BNRoleSyncCommittee,
	spectypes.BNRoleSyncCommitteeContribution,
}

func TestingStores() *aleastorage.ALEAStores {
	return aleastorage.NewStoresFromRoles(getDB(), zap.L(), allRoles...)
}
