package apiConfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	errorHandler "github.com/Aloct/DFV-Utils/internAPIUtils/errorHandling"
	"github.com/awnumar/memguard"
)

func WriteSuccessResponse(w http.ResponseWriter, status int, context string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(stdResponse{
		Context: context,
	})

	return nil
}

func WriteJSONResponse(w http.ResponseWriter, status int, context string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != "" {
		json.NewEncoder(w).Encode(stdResponse{
			Context: context,
			Data:    data,
		})
	} else {
		return errors.New("no data given")
	}

	return nil
}

func WriteError(w http.ResponseWriter, status int, err error, internalCode int, context string, addInfo string) error {
	log.Println("Error:", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if internalCode != 0 {
		errorContext, err := errorHandler.CreateHTTPError(internalCode, context, addInfo)
		if err != nil {
			http.Error(w, "failed to process request, no context given", status)
			return err
		}

		json.NewEncoder(w).Encode(errorContext)
	} else {
		http.Error(w, "failed to process request, no context given", status)
	}

	return nil

	// SEND INTO LOG

}

func SetEnclaveAsJSON(w http.ResponseWriter, key *memguard.Enclave) error {
	w.Header().Set("Content-Type", "application/json")

	lockedBuffer, err := key.Open()
	if err != nil {
		return err
	}
	defer lockedBuffer.Destroy()

	b64Encoded := base64.StdEncoding.EncodeToString(lockedBuffer.Bytes())

	json.NewEncoder(w).Encode(b64Encoded)

	return nil
}
