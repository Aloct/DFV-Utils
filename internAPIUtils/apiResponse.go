package internAPIUtils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
		Status:  status,
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