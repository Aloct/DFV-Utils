package apiConfig
// internal payload types
// 7001 - simpleType
// 7010 - keyIdentComb

type stdResponse struct {
	Context string      `json:"context"`
	Data    interface{} `json:"data,omitempty"`
}

// complex structs for Data field
type responseCreator struct{}

func NewResponseCreator() responseCreator {
	return responseCreator{}
}

type PasetoIdentifier struct {
	KEK string `json:"kek"`
	KEKDB string `json:"kekdb"`
	ID string `json:"id"`
}

func (responseCreator) NewStdResponse(context string, data interface{}) interface{} {
	return stdResponse{
		Context: context,
		Data:    data,
	}
}

func (responseCreator) NewPasetoIdentifier(kek, kekdb, id string) interface{} {
	return PasetoIdentifier{
		KEK:  kek,
		KEKDB: kekdb,
		ID: id,
	}
}