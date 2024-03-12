package dreams

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/blang/semver/v4"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/walletapi"
	"go.etcd.io/bbolt"
)

// This file is highly influenced from derohe https://github.com/deroproject/derohe
// dReams is using walletapi package to create encrypted account stores for dApp data

// Wallets signed in with a wallet file are using their internal Encrypt/Decrypt functions for a high level of data protection

// Wallets connected with RPC/XSWD are using walletapi.EncryptWithKey and DecryptWithKey functions to mirror
// the encryption scheme of the wallet files. The accounts for these connections types could be improved upon

// Like walletapi.Wallet_Memory, for storing in encrypted form
type AccountEncrypted struct {
	Version   semver.Version `json:"version"`
	pbkdf2    []byte         // used to encrypt metadata on updates
	master    []byte         // single password which never changes
	Secret    []byte         `json:"secret"`
	Encrypted []byte         `json:"encrypted"`
	KDF       walletapi.KDF  `json:"kdf"` // see this https://godoc.org/golang.org/x/crypto/pbkdf2
	account   *AccountData   // not serialized, we store an encrypted version
	sync.RWMutex
}

// Account data to store for a connected wallet
type AccountData struct {
	Tables  []string `json:"tables"`
	Predict []string `json:"predict"`
	Sports  []string `json:"sports"`
}

const (
	accountKey    = "config"
	accountBucket = "account"
)

// Account variable
var myAccount *AccountEncrypted

// Initialize myAccount variable when package is used
func init() {
	SignOut()
}

// Reset myAccount variable
func SignOut() {
	myAccount = &AccountEncrypted{
		account: &AccountData{},
	}
}

// Account address storage path, connection types are stored separately
func shardAddress() string {
	if rpc.Wallet.File != nil {
		return fmt.Sprintf("%x", sha1.Sum([]byte(rpc.Wallet.Address)))
	} else {
		return fmt.Sprintf("%x", sha1.Sum([]byte(rpc.Wallet.Address+"1")))
	}
}

// Find path for stored data
//   - 'public' true will return settings DB
//   - 'public' false will return account DB
func getShard(public bool) (db *bbolt.DB, err error) {
	var dir string
	dir, err = os.Getwd()
	if err != nil {
		return
	}

	var shard string
	if public {
		shard = "settings"
	} else {
		if rpc.Wallet.Address == "" {
			err = fmt.Errorf("nil wallet for account store")
			return
		}

		shard = shardAddress()
	}

	path := filepath.Join(dir, "datashards", shard)

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}

	db, err = bbolt.Open(filepath.Join(path, shard+".db"), 0600, nil)
	if err != nil {
		return
	}

	return
}

// Delete local storage for connected wallet
func DeleteShard() error {
	if rpc.Wallet.Address == "" {
		return fmt.Errorf("no shard address")
	}

	return os.RemoveAll(filepath.Clean(filepath.Join("datashards", shardAddress())))
}

// Store a public value in DB
func StoreValue(bucket, key string, store interface{}) (err error) {
	db, err := getShard(true)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return
		}

		mar, err := json.Marshal(&store)
		if err != nil {
			return
		}

		err = b.Put([]byte(key), mar)
		if err != nil {
			return
		}

		return
	})

	db.Close()

	return
}

// Get value from public DB
func GetValue(bucket, key string, out interface{}) (err error) {
	var db *bbolt.DB
	db, err = getShard(true)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) (err error) {
		if b := tx.Bucket([]byte(bucket)); b != nil {
			if stored := b.Get([]byte(key)); stored != nil {
				err = json.Unmarshal(stored, &out)
				if err != nil {
					return
				}

				return
			}
		}
		return
	})

	db.Close()

	return
}

// Add account data to account variable
func AddAccountData(data *AccountData) *AccountEncrypted {
	if myAccount.account == nil {
		SignOut()
	}

	myAccount.account = data

	return myAccount
}

// Store encrypted dreams account in DB
func StoreAccount(store *AccountEncrypted) (err error) {
	db, err := getShard(false)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists([]byte(accountBucket))
		if err != nil {
			return
		}

		mar, err := json.Marshal(&store)
		if err != nil {
			return
		}

		data, err := store.EncryptAccount("")
		if err != nil {
			return
		}

		myAccount.Lock()
		defer myAccount.Unlock()
		err = json.Unmarshal(data, &myAccount)
		if err != nil {
			return
		}

		err = b.Put([]byte(accountKey), data)
		if err != nil {
			return
		}

		logger.Debugln("[StoreAccount]", string(mar))

		return
	})

	db.Close()

	if err != nil {
		logger.Errorln("[StoreAccount]", err)
	}

	return
}

// Get and decrypt account dreams from DB
func GetAccount(out *AccountData) (err error) {
	var db *bbolt.DB
	db, err = getShard(false)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(accountBucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", accountBucket)
		}

		stored := b.Get([]byte(accountKey))
		if stored == nil {
			return fmt.Errorf("key %s not found", accountKey)
		}

		var data []byte
		var result *AccountData
		result, err = DecryptAccount("")
		if err != nil {
			return
		}

		data, err = json.Marshal(result)
		if err != nil {
			return
		}

		logger.Debugln("[GetAccount]", string(data))

		err = json.Unmarshal(data, &out)
		if err != nil {
			return
		}

		return
	})

	db.Close()

	return
}

// Create a new dreams account, based off walletapi.Create_Encrypted_Wallet_Memory()
func CreateAccount() (m *AccountEncrypted, err error) {
	m = &AccountEncrypted{
		Version: rpc.Version(),
		account: &AccountData{},
	}

	// generate a 64 byte key to be used as master Key
	m.master = make([]byte, 32)
	_, err = rand.Read(m.master)
	if err != nil {
		return nil, err
	}

	// walletapi (w *Wallet_Memory) Set_Encrypted_Wallet_Password()

	// set up KDF structure
	m.KDF.Salt = make([]byte, 32)
	_, err = rand.Read(m.KDF.Salt)
	if err != nil {
		return
	}
	m.KDF.Keylen = 32
	m.KDF.Iterations = 262144
	m.KDF.Hashfunction = "SHA1"

	_, err = m.EncryptAccount("")
	if err != nil {
		return
	}

	myAccount = m

	return
}

// Encrypt dreams account data, (w *Wallet_Memory) Save_Wallet()
func (m *AccountEncrypted) EncryptAccount(password string) (result []byte, err error) {
	var serialized []byte
	if rpc.Wallet.File == nil {
		if password == "" {
			// TODO something better
			password = fmt.Sprintf("%x", sha256.Sum256([]byte(rpc.Wallet.Address)))
		}

		m.pbkdf2 = walletapi.Generate_Key(m.KDF, password)

		// encrypted the master password with the pbkdf2
		m.Secret, err = walletapi.EncryptWithKey(m.pbkdf2[:], m.master)
		if err != nil {
			return
		}

		// encrypt the account
		serialized, err = json.Marshal(m.account)
		if err != nil {
			return
		}

		m.Encrypted, err = walletapi.EncryptWithKey(m.master, serialized)
		if err != nil {
			return
		}
	} else {
		serialized, err = json.Marshal(m.account)
		if err != nil {
			return
		}

		// If using wallet file its Encrypt() will be used
		m.Encrypted, err = rpc.Wallet.File.Encrypt(serialized)
		if err != nil {
			return
		}
	}

	// json marshal memory data struct, serialize it, encrypt it and store it
	result, err = json.Marshal(&m)
	if err != nil {
		return
	}

	return
}

// Decrypt dreams account data, walletapi.Open_Encrypted_Wallet_Memory()
func DecryptAccount(password string) (result *AccountData, err error) {
	w := &AccountEncrypted{}

	data, err := json.Marshal(&myAccount)
	if err != nil {
		return
	}

	// deserialize json data
	err = json.Unmarshal(data, &w)
	if err != nil {
		return
	}

	var account_bytes []byte
	if rpc.Wallet.File == nil {
		if password == "" {
			// TODO something better
			password = fmt.Sprintf("%x", sha256.Sum256([]byte(rpc.Wallet.Address)))
		}

		// try to de-seal password and store it
		w.pbkdf2 = walletapi.Generate_Key(myAccount.KDF, password)

		// try to decrypt the master password with the pbkdf2
		w.master, err = walletapi.DecryptWithKey(w.pbkdf2, w.Secret) // decrypt the master key
		if err != nil {
			return
		}

		// password has been found, open the account
		account_bytes, err = walletapi.DecryptWithKey(w.master, w.Encrypted)
		if err != nil {
			return
		}
	} else {
		// If using wallet file its Decrypt() will be used
		account_bytes, err = rpc.Wallet.File.Decrypt(w.Encrypted)
		if err != nil {
			return
		}
	}

	w.account = &AccountData{} // allocate a new instance
	err = json.Unmarshal(account_bytes, w.account)
	if err != nil {
		return
	}

	myAccount = w
	result = w.account

	return
}

// Check if account data exists in DB
func AccountExists() (found bool, account *AccountEncrypted, err error) {
	var db *bbolt.DB
	db, err = getShard(false)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountBucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", accountBucket)
		}

		value := b.Get([]byte(accountKey))
		if value != nil {
			found = true
			err = json.Unmarshal(value, &account)
			if err != nil {
				return err
			}
			return nil
		}

		return fmt.Errorf("account not found")
	})

	db.Close()

	return
}

// Create a new account if none exists
func CreateAccountIfNone() (found bool) {
	var acc *AccountEncrypted
	found, acc, _ = AccountExists()
	if !found {
		var err error
		logger.Debugln("[CreateAccountIfNone] No account, creating")
		acc, err = CreateAccount()
		if err != nil {
			logger.Errorln("[CreateAccountIfNone]", err)
			return
		}
	}

	myAccount = acc

	return
}
