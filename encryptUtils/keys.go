package encryptUtils

import (
	"crypto/rand"
	"errors"

	"github.com/awnumar/memguard"
)

type dekCombKEK struct {
	Dek *memguard.Enclave
	KekDB string
}

// dek is encrypted via the kek which is referenced by kekRef, passed dek must be unencrypted
func CreateDEKCombKEK(dek *memguard.Enclave, kekDB string) *dekCombKEK {
	return &dekCombKEK{
		Dek: dek,
		KekDB: kekDB,
	}
}

// https://developer.ibm.com/tutorials/docker-dev-db/

func (*dekCombKEK) DecryptDEK() *memguard.Enclave {
	return nil
}

func (*dekCombKEK) decryptKEK() *memguard.Enclave {
	return nil
}

func CreateAESKey(keySize int) (*memguard.Enclave, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
		return nil, errors.New("no valid key byte slice provided")
	} 

	keyRaw := make([]byte, keySize)
	if _, err := rand.Read(keyRaw); err != nil {
		return nil, errors.New("failed to create key")
	}

	key := memguard.NewEnclave(keyRaw)
	toZero(keyRaw)

	return key, nil
}

func CreateECCKey() (*memguard.Enclave, error) {
	private, err := CreateAESKey(32)
	if err != nil {
		return nil, err
	}

	return private, nil
}