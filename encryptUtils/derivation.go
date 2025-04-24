package encryptUtils

import (
	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	"golang.org/x/crypto/sha3"
)

// have user want DEK
// userBlindDEK = Hash(masterSaltSecond + scope + userRef + "DEK") => Hash(userblind + masterSaltFirst) => DEK

// have KEK want DEK
// kekBlind = Hash(masterSaltSecond + scope + kekUniqueID + "DEK") => Hash(kekBlind + masterSaltFirst) => DEK

// have user want KEK
// userBlindKEK = Hash(masterSaltFirst + scope + userRef + "KEK") => Hash(userBlind + masterSaltSecond) => KEK

// have DEK want KEK
// dekBlind = Hash(masterSaltFirst + scope + dekUniqueID + "KEK") => Hash(dekBlind + masterSaltSecond) => KEK

// System uses userReference "0"

func uniHash(data... string) []byte {
	hash := sha3.New256()
	for _, d := range data {
		hash.Write([]byte(d))
	}
	return hash.Sum([]byte{})
}

func CreateUserBlind(serviceMasterSalt, scope, userRef, wantedKeyType string) (string, error) {
	hash := uniHash(serviceMasterSalt, scope, userRef, wantedKeyType)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
} 

func CreateKeyBlind(serviceMasterSalt, scope, haveKeyUniqueID, wantedKeyType string) (string, error) {
	hash := uniHash(serviceMasterSalt, scope, haveKeyUniqueID, wantedKeyType)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
}

func HashBlind(serviceMasterSalt, blind string) (string, error) {
	hash := uniHash(serviceMasterSalt, blind)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
}