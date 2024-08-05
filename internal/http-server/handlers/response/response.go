package resp

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

// DetailedResponse represents a more detailed API response structure for errors.
type DetailedResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Code   string `json:"code,omitempty"`
	Advice string `json:"advice,omitempty"`
}

// Constants for response status and error codes.
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
)

// OK returns a response indicating a successful operation.
func OK() DetailedResponse {
	return DetailedResponse{Status: StatusOK}
}

// Error returns a detailed error response with a given message and code.
func Error(msg, code, advice string) DetailedResponse {
	return DetailedResponse{
		Status: StatusError,
		Error:  msg,
		Code:   code,
		Advice: advice,
	}
}

// ValidationError returns a detailed response indicating validation errors with detailed messages.
func ValidationError(errs validator.ValidationErrors) DetailedResponse {
	var errMsgs []string

	// Iterate over validation errors and generate error messages based on the validation tags.
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", strings.ToLower(err.Field())))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s should be a valid email", strings.ToLower(err.Field())))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", strings.ToLower(err.Field())))
		}
	}

	// Join error messages and return the response.
	return DetailedResponse{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
		Code:   CodeValidationError,
		Advice: "Check the input fields for validation errors",
	}
}
