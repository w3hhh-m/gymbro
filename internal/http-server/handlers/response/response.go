package resp

type DetailedResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error,omitempty"`
	Code   string      `json:"code,omitempty"`
	Advice string      `json:"advice,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

const (
	StatusOK            = "OK"
	StatusError         = "ERROR"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeValidationError = "VALIDATION_ERROR"
	CodeUserExists      = "USER_EXISTS"
	CodeOAuthError      = "OAUTH_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeBadRequest      = "BAD_REQUEST"
	CodeActiveWorkout   = "ACTIVE_WORKOUT"
	CodeNoActiveWorkout = "NO_ACTIVE_WORKOUT"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
)

func OK() DetailedResponse {
	return DetailedResponse{Status: StatusOK}
}

func Error(msg, code, advice string) DetailedResponse {
	return DetailedResponse{
		Status: StatusError,
		Error:  msg,
		Code:   code,
		Advice: advice,
	}
}

func Data(data interface{}) DetailedResponse {
	return DetailedResponse{
		Status: StatusOK,
		Code:   StatusOK,
		Data:   data,
	}
}
