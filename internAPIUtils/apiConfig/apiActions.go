package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/awnumar/memguard"
)

func GetEnclaveFromJSON(r *http.Request) (*memguard.Enclave, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var b64Encoded string
	err = json.Unmarshal(body, &b64Encoded)
	if err != nil {
		return nil, err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Encoded)
	if err != nil {
		return nil, err
	}

	enclave := memguard.NewEnclave(decodedBytes)

	return enclave, nil
}

func GetJSONFromResponse(r http.Response) (*stdResponse, error) {
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

func GetJSONFromRequest(r *http.Request) (*stdResponse, error) {
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