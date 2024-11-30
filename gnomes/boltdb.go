package gnomes

import (
	"encoding/json"
	"time"

	"github.com/civilware/tela/logger"
	"go.etcd.io/bbolt"
)

// Store data in boltdb, if using gravdb it will not store index
func StoreBolt(bucket, key string, store interface{}) (err error) {
	if gnomes.DBType != "boltdb" {
		logger.Debugf("[StoreBolt] DB not boltdb\n")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugf("[StoreBolt] DB is nil\n")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	for gnomes.IsWriting() {
		time.Sleep(20 * time.Millisecond)
		logger.Debugf("[StoreBolt] Write wait for %s\n", key)
	}

	gnomes.Writing(true)

	err = db.Update(func(tx *bbolt.Tx) (err error) {

		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			logger.Debugf("[StoreBolt] err creating bucket %s\n", err)
			return
		}

		mar, err := json.Marshal(&store)
		if err != nil {
			logger.Debugf("[StoreBolt] %s %v %s %s\n", key, mar, bucket, err)
			return
		}

		err = b.Put([]byte(key), []byte(mar))
		if err != nil {
			logger.Debugf("[StoreBolt] %s %v %s %s\n", key, mar, bucket, err)
			return
		}

		return
	})

	gnomes.Writing(false)

	return
}

// Get data from boltdb
func GetStorage(bucket, key string, out interface{}) {
	if gnomes.DBType != "boltdb" {
		logger.Debugf("[GetStorage] DB not boltdb\n")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugf("[GetStorage] DB is nil\n")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)); b != nil {
			if ok := b.Get([]byte(key)); ok != nil {
				err := json.Unmarshal(ok, &out)
				if err != nil {
					logger.Debugf("[GetStorage] %s\n", err)
					return err
				}
				return nil
			}
			logger.Debugf("[GetStorage] Key %s is nil\n", key)
		}
		return nil
	})
}

// Delete data from boltdb
func DeleteStorage(bucket, key string) {
	if gnomes.DBType != "boltdb" {
		logger.Debugf("[DeleteStorage] DB not boltdb\n")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugf("[DeleteStorage] DB is nil\n")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})

	if err != nil {
		logger.Debugf("[DeleteStorage] %s %s\n", bucket, err)
		return
	}

	logger.Debugf("[DeleteStorage] %s deleted\n", key)
}

// Check if data exists in boltdb
func StorageExists(bucket, key string) (found bool, err error) {
	if gnomes.DBType != "boltdb" {
		logger.Debugf("[StorageExists] DB not boltdb\n")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugf("[StorageExists] DB is nil\n")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			logger.Debugf("[StorageExists] Bucket %s not found\n", bucket)
			return nil
		}

		value := b.Get([]byte(key))
		if value == nil {
			logger.Debugf("[StorageExists] Key %s does not exist\n", key)
		} else {
			found = true
			logger.Debugf("[StorageExists] Key %s exists\n", key)
		}
		return nil
	})

	return
}
