package holdero

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/dReam-dApps/dReams/rpc"
)

// Gets local cards with local key
func Card(hash string) int {
	for i := 1; i < 53; i++ {
		finder := strconv.Itoa(i)
		add := rpc.Wallet.ClientKey + finder + Round.SC_seed
		card := sha256.Sum256([]byte(add))
		str := hex.EncodeToString(card[:])

		if str == hash {
			return i
		}

	}
	return 0
}

// Generate a new Holdero key
func GenerateKey() string {
	random, _ := rand.Prime(rand.Reader, 128)
	shasum := sha256.Sum256([]byte(random.String()))
	str := hex.EncodeToString(shasum[:])
	rpc.Wallet.KeyLock = true
	EncryptFile([]byte(str), "config/.key", rpc.Wallet.UserPass, rpc.Wallet.Address)
	log.Println("[Holdero] Round Key: ", str)
	rpc.AddLog("Round Key: " + str)

	return str
}

// Create pass hash
func createHash(key string) string {
	sha := sha256.Sum256([]byte(key))
	md5 := md5.New()
	md5.Write([]byte(hex.EncodeToString(sha[:])))
	return hex.EncodeToString(md5.Sum(nil))
}

// Encrypt plaintext data with pass
func Encrypt(data []byte, pass, add string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(pass)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("[Encrypt]", err)
		return nil
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Println("[Encrypt]", err)
		return nil
	}

	extra := []byte(add)

	return gcm.Seal(nonce, nonce, data, extra)
}

// Decrypt ciphertext with pass
func Decrypt(data []byte, pass, add string) []byte {
	key := []byte(createHash(pass))
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("[Decrypt]", err)
		return nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("[Decrypt]", err)
		return nil
	}

	extra := []byte(add)

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, extra)
	if err != nil {
		log.Println("[Decrypt]", err)
		return nil
	}

	return plaintext
}

// Write encrypted file
func EncryptFile(data []byte, filename, pass, add string) {
	if data != nil {
		if file, err := os.Create(filename); err == nil {
			defer file.Close()
			file.Write(Encrypt(data, pass, add))
		}
	}
}

// Decrypt a file
func DecryptFile(filename, pass, add string) []byte {
	if data, err := os.ReadFile(filename); err == nil {
		return Decrypt(data, pass, add)
	}
	return nil
}

// Check if Holdero key exists and decrypt
func CheckExistingKey() {
	if _, err := os.Stat("config/.key"); err == nil {
		key := DecryptFile("config/.key", rpc.Wallet.UserPass, rpc.Wallet.Address)
		if key != nil {
			rpc.Wallet.ClientKey = string(key)
			rpc.Wallet.KeyLock = true
			return
		}
	}

	shasum := sha256.Sum256([]byte("nil"))
	str := hex.EncodeToString(shasum[:])
	rpc.Wallet.ClientKey = str
}
