package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	errorHandler "github.com/Aloct/DFV-Utils/internAPIUtils/errorHandling"
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

func GetStdResponse(r http.Response) (*StdResponse, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var data StdResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, err
}

func GetErrorResponse(r http.Response) (*errorHandler.HTTPErrorContext, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var data errorHandler.HTTPErrorContext 
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, err
}