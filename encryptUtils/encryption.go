package encryptUtils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/awnumar/memguard"
)

func aesDecryption(ciphertext *memguard.Enclave, key *memguard.Enclave) (*memguard.Enclave, error) {
	keyLocked, err := key.Open()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(keyLocked.Bytes())
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

func aesEncryption(plaintext *memguard.Enclave, key *memguard.Enclave) (*memguard.Enclave, error) {
	keyLocked, err := key.Open()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(keyLocked.Bytes())
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