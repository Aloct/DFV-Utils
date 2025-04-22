package apiConfig

// internal payload types
// 7001 - simpleType
// 7010 - keyIdentComb

// complex structs for Data field
type responseCreator struct{}

func NewResponseCreator() *responseCreator {
	return &responseCreator{}
}

type stdResponse struct {
	Context string      `json:"context"`
	Data    interface{} `json:"data,omitempty"`
}

func (*responseCreator) NewStdResponse(context string, data interface{}) interface{} {
	return stdResponse{
		Context: context,
		Data:    data,
	}
}

type groupResponse struct {
	Context string      `json:"context"`
	SubResponses []interface{} `json:"subresponses"`
}

func (*responseCreator) NewGroupResponse(context string, subResponses... interface{}) interface{} {
	return groupResponse{
		Context: context,
		SubResponses: subResponses,
	}
}



type DEKRegister struct {

}

func (*responseCreator) NewDEKRegister() {
	
}

type KEKRegister struct {
	KEKDB string `json:"kekdb"`
	Scope string `json:"scope"`
	ID string `json:"id"`
	UserBlind string `json:"userblind"`
}

func (*responseCreator) NewKEKRegister(kekdb, scope, id, userBlind string) interface{} {
	return KEKRegister{
		KEKDB: kekdb,
		Scope: scope,
		ID: id,
		UserBlind: userBlind,
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