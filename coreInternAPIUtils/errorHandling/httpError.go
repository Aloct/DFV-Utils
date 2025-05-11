package errorHandler

import (
	"fmt"
	"net/http"
)

type HTTPErrorContext struct {
	Status       int
	InternalCode int
	Message      string
	Data         string
}

var loginURL = "http://localhost:8081/login"

// internal Codes:
// 6001 - login then redirect
// 6002 - internal server error, try again later
// 6003 - bad request, add ...
// 6004 - third party error, try again later
// 7001 - created

func CreateHTTPError(internalCode int, context string, addInfo string) (*HTTPErrorContext, error) {
	switch internalCode {
		case 6001:
			return &HTTPErrorContext{
				Status: 	 http.StatusUnauthorized,
				InternalCode: 6001,
				Message: 	 "Please login, " + context + "!",
				Data: 		 fmt.Sprintf("%s,%s", loginURL, addInfo),
			}, nil

		case 6002:
			return &HTTPErrorContext{
				Status: 	 http.StatusInternalServerError,
				InternalCode: 6002,
				Message: 	 "Please try again later",
				Data: 		 "",
			}, nil

		case 6003:
			return &HTTPErrorContext{
				Status: 	 http.StatusBadRequest,
				InternalCode: 6003,
				Message: 	 fmt.Sprintf("Bad request, no valid %s", context),
				Data: 		 "",
			}, nil

		case 6004:
			return &HTTPErrorContext{
				Status: 	 http.StatusFailedDependency,
				InternalCode: 6004,
				Message: 	 "Please try again later, third party error",
				Data: 		 "",
			}, nil
	}

	return nil, fmt.Errorf("invalid internal code")
}