package internAPIUtils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/awnumar/memguard"
)

type stdResponse struct {
	Status  int         `json:"status"`
	Context string      `json:"context"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJson(w http.ResponseWriter, status int, context string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(stdResponse{
		Context: context,
		Data:    data,
	})

	return nil
}

func WriteError(w http.ResponseWriter, status int, err error) {
	fmt.Fprintln(os.Stderr, err)
	http.Error(w, err.Error(), status)

	// SEND INTO LOG
}

func SetEnclaveAsJSON(w http.ResponseWriter, key *memguard.Enclave, add string) error {
	w.Header().Set("Content-Type", "application/json")

	lockedBuffer, err  := key.Open()
	if err != nil {
		return err
	}
	defer lockedBuffer.Destroy()

	b64Encoded := base64.StdEncoding.EncodeToString(lockedBuffer.Bytes())

	b64Encoded = b64Encoded + ";" + add

	json.NewEncoder(w).Encode(b64Encoded)

	return nil
}