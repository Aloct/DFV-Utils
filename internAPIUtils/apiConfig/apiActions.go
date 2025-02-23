package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/awnumar/memguard"
)

func GetEnclaveFromJSON(r *http.Request) (*memguard.Enclave, string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}
	defer r.Body.Close()

	var b64Encoded string
	err = json.Unmarshal(body, &b64Encoded)
	if err != nil {
		return nil, "", err
	}

	splitted := strings.Split(b64Encoded, ";")

	keyPart, add := splitted[0], splitted[1]

	decodedBytes, err := base64.StdEncoding.DecodeString(keyPart)
	if err != nil {
		return nil, "", err
	}

	enclave := memguard.NewEnclave(decodedBytes)

	return enclave, add, nil
}

func GetStdResponse(r http.Response) (*stdResponse, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var data stdResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, err
}