package encryptionHelper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/awnumar/memguard"
	ecies "github.com/ecies/go"
)

type keyDepManager struct {
	symmetricKeyDep string
	asymmetricSecretDep string

	optKeyDep string

	implicit    []byte
	keyLifetime time.Duration

	fileDir string
	dbPath string
}

// encryptionType => data-master.KEK.DEK
func validateEncryptionType(givenType string) (*[]string, error) {
	splitted := strings.Split(givenType, ".")

	if (len(splitted) != 3) {
		return nil, errors.New("No valid encryptionType given, too short")
	} 

	return *splitted, nil
}

func initKeyDepManager(implicit string, pasetoKeys bool, encryptionType string) (error, *keyDepManager) {
	if (!symmetric && !asymmetric) {
		return errors.New("Key Manager must consist of at least one key type"), nil
	}

	var keyHolder map[string]*memguard.Enclave = make(map[string]*memguard.Enclave)

	splittedType := &validateEncryptionType(encryptionType)
	encryptionRdyType := ""

	curKeyComb := ""
	for i := 1; i < len(splittedType); i++ {
		asymmetricCounter := 1
		symmetricCounter := 1

		if (splittedType[i] == "as") {
			curKeyComb = "as" + fmt.Sprintf("-%s", asymmetricCounter)

			if (pasetoKeys) {
				keyHolder[curKeyComb] = memguard.NewEnclave(paseto.NewV4AsymmetricSecretKey().ExportBytes())
			} else {
				key, err := ecies.GenerateKey()
				if err != nil {
					fmt.Println(err)
					return err, nil
				}
	
				keyHolder[curKeyComb] = memguard.NewEnclave(key.Bytes())
			}

			asymmetricCounter++
		} else if (splittedType[i] == "s") {
			curKeyComb = "s" + fmt.Sprintf("-%s", symmetricCounter)	
	
			keyHolder[curKeyComb] = memguard.NewEnclaveRandom(32)

			symmetricCounter++
		} else {
			return errors.New(fmt.Sprintf("Invalid encryption type at index %d: %s", i, splittedType[i])), nil
		}
		
		encryptionRdyType := encryptionRdyType + curKeyComb
	}

	// enrcyptionType => data-master.KEK.DEK
	var keyDeps map[string]string = cryptoHelper.EncryptAndStoreKeys(keyHolder, encryptionRdyType)
	keyHolder = make(map[string]*memguard.Enclave)

	return nil, &keyDepManager{[]byte(implicit), time.Hour * 24}
}