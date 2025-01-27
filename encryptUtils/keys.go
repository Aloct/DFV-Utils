package encryptUtils

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"errors"

	"github.com/awnumar/memguard"
)

type dekCombKEK struct {
	Dek []byte
	KekDB *sql.DB
}

// dek must be encrypted via the kek which is referenced by kekRef
func CreateDEKCombKEK(dek []byte, kekDB *sql.DB) *dekCombKEK {
	return &dekCombKEK{
		Dek: dek,
		KekDB: kekDB,
	}
}

func (*dekCombKEK) DecryptDEK() *memguard.Enclave {

}

func (*dekCombKEK) decryptKEK() *memguard.Enclave {

}

func CreateECCKey() (ed25519.PrivateKey, error) {
	_, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return private, nil
}

func CreateAESKey(keySize int) ([]byte, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
		return nil, errors.New("no valid key byte slice provided")
	} 

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, errors.New("failed to create key")
	}

	return key, nil
}