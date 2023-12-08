package gnomes

import (
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"
)

// Store data in boltdb, if using gravdb it will not store index
func StoreBolt(bucket, key string, store interface{}) (err error) {
	if gnomes.DBType != "boltdb" {
		logger.Debugln("[StoreBolt] DB not boltdb")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugln("[StoreBolt] DB is nil")
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
			logger.Debugln("[StoreBolt]", key, mar, bucket, err)
			return
		}

		err = b.Put([]byte(key), []byte(mar))
		if err != nil {
			logger.Debugln("[StoreBolt]", key, mar, bucket, err)
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
		logger.Debugln("[GetStorage] DB not boltdb")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugln("[GetStorage] DB is nil")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)); b != nil {
			if ok := b.Get([]byte(key)); ok != nil {
				err := json.Unmarshal(ok, &out)
				if err != nil {
					logger.Debugln("[GetStorage]", err)
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
		logger.Debugln("[DeleteStorage] DB not boltdb")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugln("[DeleteStorage] DB is nil")
		return
	}

	db := gnomes.Indexer.BBSBackend.DB
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})

	if err != nil {
		logger.Debugln("[DeleteStorage]", bucket, err)
		return
	}

	logger.Debugln("[DeleteStorage]", key, "deleted")
}

// Check if data exists in boltdb
func StorageExists(bucket, key string) (found bool, err error) {
	if gnomes.DBType != "boltdb" {
		logger.Debugln("[StorageExists] DB not boltdb")
		return
	}

	if gnomes.Indexer == nil {
		logger.Debugln("[StorageExists] DB is nil")
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
