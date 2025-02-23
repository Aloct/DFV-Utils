package encryptUtils

import (
	"encoding/hex"
	"reflect"
)

func toZero(byteSlice []byte) {
	for i := range byteSlice {
		byteSlice[i] = 0
	}
}

func KeyToString(key []byte) string {
	return hex.EncodeToString(key)
}

func StringToKey(keyRaw string) ([]byte, error) {
	var keyString string

	if (reflect.TypeOf(keyRaw).Kind() == reflect.String) {
		keyString = keyRaw
	} else if (reflect.TypeOf(keyRaw).Kind() == reflect.Slice) { 
		keyString = string(keyRaw)
	}

	decodedKey, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	return decodedKey, err
}