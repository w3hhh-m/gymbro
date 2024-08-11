package validation

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"strings"
)

func ValidationError(errs validator.ValidationErrors) resp.DetailedResponse {
	var errMsgs []string

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

	return resp.DetailedResponse{
		Status: resp.StatusError,
		Error:  strings.Join(errMsgs, ", "),
		Code:   resp.CodeValidationError,
		Advice: "Check the input fields for validation errors",
	}
}

func ValidateStruct(log *slog.Logger, s interface{}) error {
	if err := validator.New().Struct(s); err != nil {
		log.Debug("Failed to validate request", slog.Any("error", err))
		return err
	}
	return nil
}

func HandleValidationError(w http.ResponseWriter, r *http.Request, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ValidationError(ve))
	}
}
