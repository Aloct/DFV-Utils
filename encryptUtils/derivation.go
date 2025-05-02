package encryptUtils

import (
	"crypto/hmac"
	"crypto/sha256"

	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	"golang.org/x/crypto/sha3"
)

// have user want DEK
// user => KEK => DEK

// have KEK want DEK
// kekBlind = Hash(masterSaltSecond + innerScope + KEKHash + "DEK") => Hash(kekBlind + masterSaltFirst) => DEK

// have user want KEK
// userBlindKEK = Hash(masterSaltFirst + scope + userRef + "KEK") => Hash(userBlind + masterSaltSecond) => KEK

// System uses userReference "0"

func uniHash(data... string) []byte {
	hash := sha3.New256()
	for _, d := range data {
		hash.Write([]byte(d))
	}
	return hash.Sum([]byte{})
}

func uniHMAC(secret []byte, data ...string) []byte {
	mac := hmac.New(sha256.New, secret)
	for _, d := range data {
		mac.Write([]byte(d))
	}
	return mac.Sum(nil)
}

func CreateUserBlind(serviceHashKey []byte, scope, userRef, wantedKeyType string) (string, error) {
	hash := uniHMAC(serviceHashKey, scope, userRef, wantedKeyType)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
} 

func CreateKeyBlind(serviceHashKey []byte, scope, KEKHash, wantedKeyType string) (string, error) {
	hash := uniHMAC(serviceHashKey, scope, KEKHash, wantedKeyType)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
}

func HashBlind(serviceHashKey []byte, blind string) (string, error) {
	hash := uniHMAC(serviceHashKey, blind)

	serialized, err := coreutils.HashToString(hash)
	if err != nil {
		return "", err
	}

	return serialized, nil
}

func HashKey(key []byte) (string, error) {
	keyString, err := coreutils.KeyToString(key)
	if err != nil {
		return "", err
	}

	hashed := uniHash(keyString)
	serialized, err := coreutils.HashToString(hashed)
	if err != nil {
		return "", err
	}

	return serialized, nil
}