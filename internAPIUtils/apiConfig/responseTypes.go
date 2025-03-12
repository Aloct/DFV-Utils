package apiConfig

import (
	"net/http"
)

// internal payload types
// 7001 - simpleType
// 7010 - keyIdentComb

type stdResponse struct {
	Context string      `json:"context"`
	Data    interface{} `json:"data,omitempty"`
}

// complex structs for Data field
type PasetoIdentifier struct {
	KEK string `json:"kek"`
	KEKDB string `json:"kekdb"`
	ID string `json:"id"`
}

func NewPasetoIdentifier(w http.ResponseWriter, kek string, kekdb, id, userref string) PasetoIdentifier {
	w.Header().Set("internalDataCode", "7010")

	return PasetoIdentifier{
		KEK:  kek,
		KEKDB: kekdb,
		ID: id,
	}
}