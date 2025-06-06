package coreutils

import (
	"encoding/hex"
	"fmt"
	"reflect"
)

func KeyToString(keyRaw any) (string, error) {
	var keyBytes []byte

	if reflect.TypeOf(keyRaw).Kind() == reflect.Slice {
		keyBytes = keyRaw.([]byte)
	} else if reflect.TypeOf(keyRaw).Kind() == reflect.String {
		keyBytes = []byte(keyRaw.(string))
	} else {
		return "", fmt.Errorf("invalid key type")
	}

	return hex.EncodeToString(keyBytes), nil
}

func StringToKey(keyRaw any) ([]byte, error) {
	var keyString string

	if reflect.TypeOf(keyRaw).Kind() == reflect.String {
		keyString = keyRaw.(string)
	} else if reflect.TypeOf(keyRaw).Kind() == reflect.Slice {
		keyString = string(keyRaw.([]byte))
	} else {
		return nil, fmt.Errorf("invalid key type")
	}

	decodedKey, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	return decodedKey, err
}

func HashToString(hashRaw any) (string, error) {
	serialized, err := KeyToString(hashRaw)
	if err != nil {
		return "", err
	}

	return serialized, nil
}