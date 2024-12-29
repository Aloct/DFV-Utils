package internAPIUtils

import (
	"encoding/json"
	"net/http"
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

func writeError(w http.ResponseWriter, status int, err error) {
	http.Error(w, err.Error(), status)

	// SEND INTO LOG
}