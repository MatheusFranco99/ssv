package storage

import (
	"fmt"
	"github.com/MatheusFranco99/ssv/storage/basedb"
	"github.com/MatheusFranco99/ssv/storage/kv"
)

// GetStorageFactory resolve and returns db instance based on db type
func GetStorageFactory(options basedb.Options) (basedb.IDb, error) {
	switch options.Type {
	case "badger-db":
		db, err := kv.New(options)
		return db, err
	case "badger-memory":
		db, err := kv.New(options)
		return db, err
	}
	return nil, fmt.Errorf("unsupported storage type passed")
}
