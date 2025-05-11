package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

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

func GetJSONFromResponse(r http.Response) (*StdResponse, error) {
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

func GetJSONFromRequest(r *http.Request) (*StdRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var data StdRequest
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
// func ProcessMultipartResponse(resp *http.Response, handlers map[string]PartHandler) (map[string]interface{}, error) {
// 	// Für Responses brauchen wir keinen ParseMultipartForm Aufruf
	
// 	// MultipartReader direkt aus der Response erstellen
// 	contentType := resp.Header.Get("Content-Type")
// 	if contentType == "" {
// 		return nil, fmt.Errorf("no Content-Type header found")
// 	}
	
// 	mediaType, params, err := mime.ParseMediaType(contentType)
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	if !strings.HasPrefix(mediaType, "multipart/") {
// 		return nil, fmt.Errorf("not a multipart response: %s", mediaType)
// 	}
	
// 	boundary, ok := params["boundary"]
// 	if !ok {
// 		return nil, fmt.Errorf("no boundary parameter found in Content-Type")
// 	}
	
// 	reader := multipart.NewReader(resp.Body, boundary)

// 	results, err := multiPartLoop(reader, handlers)
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	return results, nil
// }  

// func ProcessMultipartRequest(r *http.Request, handlers map[string]PartHandler) (map[string]interface{}, error) {
// 	// MultipartReader direkt aus der Request erstellen
// 	contentType := r.Header.Get("Content-Type")
// 	if contentType == "" {
// 		return nil, fmt.Errorf("no Content-Type header found")
// 	}
	
// 	mediaType, params, err := mime.ParseMediaType(contentType)
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	if !strings.HasPrefix(mediaType, "multipart/") {
// 		return nil, fmt.Errorf("not a multipart request: %s", mediaType)
// 	}
	
// 	boundary, ok := params["boundary"]
// 	if !ok {
// 		return nil, fmt.Errorf("no boundary parameter found in Content-Type")
// 	}
	
// 	reader := multipart.NewReader(r.Body, boundary)

// 	results, err := multiPartLoop(reader, handlers)
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	return results, nil
// }

// func multiPartLoop(reader *multipart.Reader, handlers map[string]PartHandler) (map[string]interface{}, error) {
// 	results := make(map[string]interface{})

// 	for {
// 		part, err := reader.NextPart()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return nil, err
// 		}
		
// 		formName := part.FormName()
// 		handler, exists := handlers[formName]
		
// 		if exists {
// 			result, err := handler(part)
// 			if err != nil {
// 				return nil, err
// 			}
// 			results[formName] = result
// 		}
// 	}
	
// 	// Check if all required handlers were applied
// 	for name := range handlers {
// 		if _, exists := results[name]; !exists {
// 			if name == "subResponses" { // Umbenannt von subRequests zu subResponses
// 				results[name] = nil
// 			} else {
// 				return nil, fmt.Errorf("missing required part: %s", name)
// 			}
// 		}
// 	}

// 	return results, nil
// }

// func NewHandlerMap(expectedData ...string) map[string]PartHandler {
// 	handlers := make(map[string]PartHandler)
// 	for _, data := range expectedData {
// 		switch data {
// 		case "keyMetadata":
// 			handlers[data] = MetadataHandler
// 		case "key":
// 			handlers[data] = KeyHandler
// 		case "subRequests":
// 			handlers[data] = SubRequestHandler
// 		default:
// 			panic(fmt.Sprintf("unknown part type: %s", data))
// 		}
// 	}
// 	return handlers
// }

// // returns json
// func MetadataHandler(part io.Reader) (interface{}, error) {
// 	metadataBytes, err := io.ReadAll(part)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// var metadata interface{}
// 	// err = json.Unmarshal(metadataBytes, &metadata)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return metadataBytes, nil
// }

// func KeyHandler(part io.Reader) (interface{}, error) {
// 	keyData, err := io.ReadAll(part)
// 	if err != nil {
// 		return nil, err
// 	}

// 	enclave := memguard.NewEnclave(keyData)
// 	coreutils.ToZero(keyData)

// 	return enclave, nil
// }

// func SubRequestHandler(part io.Reader) (interface{}, error) {
// 	subRequestBytes, err := io.ReadAll(part)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var subRequest GroupRequest
// 	err = json.Unmarshal(subRequestBytes, &subRequest)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return subRequest, nil
// }
