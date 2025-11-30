package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Status  int               `json:"status"`
    Details map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code, message string, status int) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Status:  status,
    }
}

// Factories (no shared instances)

func NotFound(entity string) *AppError {
    return New("NOT_FOUND", fmt.Sprintf("%s not found", entity), http.StatusNotFound)
}

func BadRequest(msg string) *AppError {
    return New("BAD_REQUEST", msg, http.StatusBadRequest)
}

func Unauthorized(msg string) *AppError {
    return New("UNAUTHORIZED", msg, http.StatusUnauthorized)
}

func Validation(details map[string]string) *AppError {
    return &AppError{
        Code:    "VALIDATION_ERROR",
        Message: "Invalid fields",
        Status:  http.StatusUnprocessableEntity,
        Details: details,
    }
}

func Internal(msg string) *AppError {
	return New("INTERNAL_ERROR", msg, http.StatusInternalServerError)
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}