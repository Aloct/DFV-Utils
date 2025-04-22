package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	errorHandler "github.com/Aloct/DFV-Utils/internAPIUtils/errorHandling"
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

// mulipart requests
func GetKeyMetaFromMultipartReq(r *http.Request, maxMemory int64) (interface{}, *memguard.Enclave, error) {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return nil, nil, err
	}

	keyMeta := r.MultipartForm.Value["keymeta"]
	if len(keyMeta) == 0 {
		return nil, nil, errors.New("no key meta data found in request")
	}

	key := r.MultipartForm.File["key"]
	if len(key) == 0 {
		return nil, nil, errors.New("no key found in request")
	} 

	

	return keyMeta, nil
}