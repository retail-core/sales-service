package validation

import (
	"fmt"
	"net/http"
	"github.com/retail-core/sales-service/internal/errors"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) *errors.AppError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &errors.AppError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request parameters",
			Status:  http.StatusUnprocessableEntity,
		}
	}

	errorsMap := make(map[string]string, len(validationErrs))
	for _, e := range validationErrs {
		errorsMap[e.Field()] = buildValidationMessage(e)
	}

	return &errors.AppError{
		Code:    "VALIDATION_ERROR",
		Message: "Invalid request parameters",
		Status:  http.StatusUnprocessableEntity,
		Details: errorsMap,
	}
}

func buildValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s is invalid", e.Field())
	}
}
