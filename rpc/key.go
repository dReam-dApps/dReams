package rpc

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
)

func Card(hash string) int { /// Gets local cards with local key
	for i := 1; i < 53; i++ {
		finder := strconv.Itoa(i)
		add := Wallet.ClientKey + finder + Round.SC_seed
		card := sha256.Sum256([]byte(add))
		str := hex.EncodeToString(card[:])

		if str == hash {
			return i
		}

	}
	return 0
}

func GenerateKey() string {
	random, _ := rand.Prime(rand.Reader, 128)
	shasum := sha256.Sum256([]byte(random.String()))
	str := hex.EncodeToString(shasum[:])
	Wallet.KeyLock = true
	EncryptFile([]byte(str), "config/.key", Wallet.UserPass, Wallet.Address)
	log.Println("[Holdero] Round Key: ", str)
	AddLog("Round Key: " + str)

	return str
}

func createHash(key string) string {
	sha := sha256.Sum256([]byte(key))
	md5 := md5.New()
	md5.Write([]byte(hex.EncodeToString(sha[:])))
	return hex.EncodeToString(md5.Sum(nil))
}

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

func EncryptFile(data []byte, filename, pass, add string) {
	if data != nil {
		if file, err := os.Create(filename); err == nil {
			defer file.Close()
			file.Write(Encrypt(data, pass, add))
		}
	}
}

func DecryptFile(filename, pass, add string) []byte {
	if data, err := os.ReadFile(filename); err == nil {
		return Decrypt(data, pass, add)
	}
	return nil
}

func CheckExisitingKey() {
	if _, err := os.Stat("config/.key"); err == nil {
		key := DecryptFile("config/.key", Wallet.UserPass, Wallet.Address)
		if key != nil {
			Wallet.ClientKey = string(key)
			Wallet.KeyLock = true
			return
		}
	}

	shasum := sha256.Sum256([]byte("nil"))
	str := hex.EncodeToString(shasum[:])
	Wallet.ClientKey = str
}
