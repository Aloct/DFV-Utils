package apiConfig

// internal payload types
// 7001 - simpleType
// 7010 - keyIdentComb

// complex structs for Data field
type message struct {
	Context string      `json:"context"`
	Data    interface{} `json:"data,omitempty"`
}

type StdResponse = message
type StdRequest = message

func NewStdResponse(context string, data interface{}) StdResponse {
	return message{
		Context: context,
		Data:    data,
	}
}

func NewStdRequest(context string, data interface{}) StdRequest {
	return message{
		Context: context,
		Data:    data,
	}
}

type groupMessage struct {
	Context string        `json:"context"`
	Subs    []interface{} `json:"subs"`
}

type GroupResponse = groupMessage
type GroupRequest = groupMessage

func NewGroupResponse(context string, subResponses ...interface{}) interface{} {
	return GroupResponse{
		Context: context,
		Subs:    subResponses,
	}
}

func NewGroupRequest(context string, subRequests ...interface{}) interface{} {
	return GroupRequest{
		Context: context,
		Subs:    subRequests,
	}
}

type DEKRegister struct {
	ID       string `json:"id"`
	Scope    string `json:"scope"`
}

func NewDEKRegister(id, scope, dekBlind string) DEKRegister {
	return DEKRegister{
		ID:       id,
		Scope:    scope,
	}
}

type KEKRegister struct {
	KEKDB     string `json:"kekdb"`
	Scope     string `json:"scope"`
	ID        string `json:"id"`
	UserBlind string `json:"userblind"`
}

func NewKEKRegister(kekdb, scope, id, userBlind string) KEKRegister {
	return KEKRegister{
		KEKDB:     kekdb,
		Scope:     scope,
		ID:        id,
		UserBlind: userBlind,
	}
}

type KEKIdentifier struct {
	KEKDB     string `json:"kekdb"`
	UserBlind string `json:"userblind"`
}

func NewKEKIdentifier(kekdb, userblind string) KEKIdentifier {
	return KEKIdentifier{
		KEKDB:     kekdb,
		UserBlind: userblind,
	}
}

type DEKIdentifier struct {
	ID       string `json:"id"`
	KEKBlind string `json:"kekblind"`
}

func NewDEKIdentifier(id, kekBlind string) DEKIdentifier {
	return DEKIdentifier{
		ID:       id,
		KEKBlind: kekBlind,
	}
}

type DEKKEKRegisterSet struct {
	DEKRegister `json:"dekset"`
	KEKIdentifier `json:"kekset"`
}

func NewDEKKEKSet(dekRegister DEKRegister, kekIdentifier KEKIdentifier) DEKKEKRegisterSet {
	return DEKKEKRegisterSet{
		DEKRegister: dekRegister,
		KEKIdentifier: kekIdentifier,
	}
}

type PasetoIdentifier struct {
	APaseto     string `json:"apaseto"`
	RPaseto     string `json:"rpaseto"`
	Fingerprint string `json:"fingerprint"`
}

func NewPasetoIdentifier(apaseto, rpaseto string) PasetoIdentifier {
	return PasetoIdentifier{
		APaseto: apaseto,
		RPaseto: rpaseto,
	}
}
