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

func NewResponseCreator() *responseCreator {
	return &responseCreator{}
}

func (*responseCreator) NewStdResponse(context string, data interface{}) interface{} {
	return stdResponse{
		Context: context,
		Data:    data,
	}
}

type KEKRegister struct {
	KEK string `json:"kek"`
	KEKDB string `json:"kekdb"`
	Scope string `json:"scope"`
	ID string `json:"id"`
	UserBlind string `json:"userblind"`
	KeyBlind string `json:"keyblind"`
}

func (*responseCreator) NewKEKRegister(kek, kekdb, scope, id, userBlind, keyBlind string) interface{} {
	return KEKRegister{
		KEK:  kek,
		KEKDB: kekdb,
		Scope: scope,
		ID: id,
		UserBlind: userBlind,
		KeyBlind: keyBlind,
	}
}

type KEKIdentifier struct {
	KEKDB string `json:"kekdb"`
	ID string `json:"id"`
	IDType string `json:"idtype"`
}

func (*responseCreator) NewKEKIdentifier(kekdb, id, idType string) interface{} {
	return KEKIdentifier{
		KEKDB: kekdb,
		ID: id,
		IDType: idType,
	}
}

type PasetoIdentifier struct {
	APaseto string `json:"apaseto"`
	RPaseto string `json:"rpaseto"`
	Fingerprint string `json:"fingerprint"`
}

func (*responseCreator) NewPasetoIdentifier(apaseto, rpaseto string) interface{} {
	return PasetoIdentifier{
		APaseto: apaseto,
		RPaseto: rpaseto,
	}
}