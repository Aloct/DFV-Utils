package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	errorHandler "github.com/Aloct/DFV-Utils/internAPIUtils/errorHandling"
	"github.com/awnumar/memguard"
)

type PartHandler func(io.Reader) (interface{}, error)

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
func ProcessMultipartRequest(r *http.Request, maxMemoryInMB int64, handlers map[string]PartHandler) (map[string]interface{}, error) {
    if err := r.ParseMultipartForm(maxMemoryInMB << 20); err != nil {
        return nil, err
    }

    reader, err := r.MultipartReader()
    if err != nil {
        return nil, err
    }

    results := make(map[string]interface{})

    for {
        part, err := reader.NextPart()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        formName := part.FormName()
        handler, exists := handlers[formName]
        
        if exists {
            result, err := handler(part)
            if err != nil {
                return nil, err
            }
            results[formName] = result
        }
    }

    // Check if all required handlers were applied
    for name := range handlers {
        if _, exists := results[name]; !exists {
            return nil, fmt.Errorf("missing required part: %s", name)
        }
    }

    return results, nil
}

func NewHandlerMap(expectedData... string) map[string]PartHandler {
	handlers := make(map[string]PartHandler)
	for _, data := range expectedData {
		switch data {
		case "metadata":
			handlers[data] = MetadataHandler
		case "key":
			handlers[data] = KeyHandler
		case "subRequest":
			handlers[data] = SubRequestHandler
		default:
			panic(fmt.Sprintf("unknown part type: %s", data))
		}
	}
	return handlers
}

func MetadataHandler(part io.Reader) (interface{}, error) {
    metadataBytes, err := io.ReadAll(part)
    if err != nil {
        return nil, err
    }

    var metadata interface{}
    err = json.Unmarshal(metadataBytes, &metadata)
    if err != nil {
        return nil, err
    }
	coreutils.ToZero(metadataBytes)
    
    return metadata, nil
}

func KeyHandler(part io.Reader) (interface{}, error) {
    keyData, err := io.ReadAll(part)
    if err != nil {
        return nil, err
    }
    
    enclave := memguard.NewEnclave(keyData)
    coreutils.ToZero(keyData)
    
    return enclave, nil
}

func SubRequestHandler(part io.Reader) (interface{}, error) {
	subRequestBytes, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	var subRequest interface{}
	err = json.Unmarshal(subRequestBytes, &subRequest)
	if err != nil {
		return nil, err
	}
	
	return subRequest, nil
}