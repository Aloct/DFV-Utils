package encryptUtils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/awnumar/memguard"
)

type dekCombKEK struct {
	Dek *memguard.Enclave
	KekDB string
}

// dek is encrypted via the kek which is referenced by kekRef, passed dek must be unencrypted
func CreateDEKCombKEK(dek *memguard.Enclave, kekDB string, managerC string, managerP int) (*dekCombKEK, error) {
	kek, err := CreateAESKey(32)
	if err != nil {
		return nil, err
	}

	dekEncrypted, err := AesEncryption(kek, dek)
	if err != nil {
		return nil, err
	}

	keyBuf, err := kek.Open()
	if err != nil {
		return nil, err
	}
	defer keyBuf.Destroy()

	jsonData, err := json.Marshal(base64.StdEncoding.EncodeToString(keyBuf.Bytes()) + ";" + kekDB)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%d/registerKEK", managerC, managerP), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to add key comb to auth manager")
	}

	return &dekCombKEK{
		Dek: dekEncrypted,
		KekDB: kekDB,
	}, nil
}

// https://developer.ibm.com/tutorials/docker-dev-db/

func (*dekCombKEK) DecryptDEK(plaintext *memguard.Enclave, encryption bool) *memguard.Enclave {
	// get KEK from decryptKEK()
	
	// decrypt DEK 

	// perfrom en/decryption with DEK
	if encryption {

	} else if !encryption {

	}

	return nil
}

func (*dekCombKEK) decryptKEK() *memguard.Enclave {
	// check if KEK is cached

	// y => fetch from redis 
	// n => fetch from sql 

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