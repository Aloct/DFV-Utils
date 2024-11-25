package encryptionHelper

import (
	"fmt"
	"os"
	"strings"

	"github.com/awnumar/memguard"
	ecies "github.com/ecies/go"
)

func eciesEncryption(key *memguard.Enclave, KEK *memguard.Enclave) (*memguard.Enclave, error) {
	lockedBufferKey, err := key.Open()
	if err != nil {
		return nil, err
	}

	lockedBufferKEK, err := KEK.Open()
	if err != nil {
		return nil, err
	}

	publicKEK, err := ecies.NewPublicKeyFromBytes(lockedBufferKEK.Bytes())
	if err != nil {
		return nil, err
	}

	encryptedK, err := ecies.Encrypt(publicKEK, lockedBufferKey.Bytes())
	if err != nil {
		return nil, err
	}

	return memguard.NewEnclave(encryptedK), nil
}

// keyType => master.  KEK.   DEK
// 			  paseto. as-1.  as-2
func EncryptKeys(keys map[string]*memguard.Enclave, eType string) map[string]*memguard.Enclave {
	var encryptMap = make(map[string]*memguard.Enclave)

	splittedType := strings.Split(eType, ".")

	var curKEK *memguard.Enclave
	for i := len(splittedType)-1; i < 0; i-- {
		curKeyType := splittedType[i]
		curKEKRef := splittedType[len(splittedType)-1-i]

		if (curKEKRef == "paseto") {
			curKEK = getPasetoMaster()
		} else if (curKEKRef == "data") {
			curKEK = getDataMaster()
		} else {
			curKEK = keys[curKEKRef]
		}


		encryptedKey, err := eciesEncryption(keys[curKeyType], curKEK)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		encryptMap[curKeyType] = encryptedKey
	}

	return encryptMap
}