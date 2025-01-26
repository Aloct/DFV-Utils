package encryptUtils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/awnumar/memguard"
)

// IMPLEMENT KMS USAGE TO PERFROM MASTER ENCRYPTION
var (
	pasetoAccessMaster = []byte("01234567890123456789012345678901")
	pasetoRefreshMaster = []byte("01234567890123456789012345678901")
)

func toZero(byteSlice []byte) {
	for i := range byteSlice {
		byteSlice[i] = 0
	}
}

func aesDecryption(ciphertext *memguard.Enclave, key []byte) (*memguard.Enclave, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	ciphertextLocked, err := ciphertext.Open()
	if err != nil {
		return nil, err
	}

	ciphertextBytes := ciphertextLocked.Bytes()
	if len(ciphertextBytes) < nonceSize {
		return nil, err
	}
	nonce, ciphertextRaw := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	ciphertextLocked.Destroy()

	plaintext, err := aesGCM.Open(nil, nonce, ciphertextRaw, nil)
	if err != nil {
		return nil, err
	}
	toZero(ciphertextRaw)

	plaintextEnclave := memguard.NewEnclave(plaintext)

	return plaintextEnclave, nil
}

func aesEncryption(plaintext *memguard.Enclave, key []byte) (*memguard.Enclave, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err 
	}

	plaintextLocked, err := plaintext.Open()
	if err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintextLocked.Bytes()), nil)
	plaintextLocked.Destroy()

	return memguard.NewEnclave(ciphertext), nil
}

func KekDecryption(kek *memguard.Enclave, master string) (*memguard.Enclave, error) {
	var masterKey []byte
	if master == "pasetoA" {
		masterKey = pasetoAccessMaster
	} else if master == "pasetoR" {
		masterKey = pasetoRefreshMaster
	}
	
	encryptedKek, err := aesDecryption(kek, masterKey)
	if err != nil {
		return nil, err
	}

	return encryptedKek, nil
}

func KekEncryption(kek *memguard.Enclave) {
}