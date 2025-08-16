package config

import (
	"errors"
	"net/http"
)

var (
	ErrCustomerNotFound       = errors.New("customer not found")
	ErrCustomerAlreadyExists  = errors.New("customer already exists")
	ErrInvalidCustomerDetails = errors.New("invalid customer details")

	// Transaction errors
	ErrTransactionNotFound       = errors.New("transaction not found")
	ErrTransactionFailed         = errors.New("transaction failed")
	ErrInvalidTransactionPayload = errors.New("invalid transaction payload")

	// Validation errors
	ErrValidationFailed      = errors.New("customer validation failed")
	ErrNoValidationLogsFound = errors.New("no validation logs found")
	ErrInvalidUploaderID     = errors.New("invalid uploader ID")

	// Rating errors
	ErrRatingNotFound        = errors.New("rating not found")
	ErrCannotCalculateRating = errors.New("unable to calculate rating")

	// Authorization / Auth errors
	ErrUnauthorizedAccess    = errors.New("unauthorized access")
	ErrTokenExpired          = errors.New("token expired or invalid")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInsufficientPrivilege = errors.New("insufficient privileges")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidCreds          = errors.New("invalid credentials")
	ErrInviteNotFound        = errors.New("invite not found")
	ErrInviteExpired         = errors.New("invite expired")
	ErrInviteAccepted        = errors.New("invite already accepted")
	// Generic / HTTP errors
	ErrBadRequest      = errors.New("bad request")
	ErrInternalServer  = errors.New("internal server error")
	ErrConflict        = errors.New("conflict detected")
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many requests")
)

func GetStatusCode(err error) int {
	switch err {
	case nil:
		return http.StatusOK

	// Bad request errors
	case ErrInvalidCustomerDetails, ErrInvalidTransactionPayload, ErrBadRequest:
		return http.StatusBadRequest

	// Unauthorized errors
	case ErrUnauthorizedAccess, ErrTokenExpired, ErrInvalidCredentials, ErrInsufficientPrivilege:
		return http.StatusUnauthorized

	// Conflict errors
	case ErrCustomerAlreadyExists, ErrConflict:
		return http.StatusConflict

	// Not found errors
	case ErrCustomerNotFound, ErrTransactionNotFound, ErrRatingNotFound, ErrNoValidationLogsFound, ErrNotFound:
		return http.StatusNotFound
	case ErrTooManyRequests:
		return http.StatusTooManyRequests

	// Unprocessable / domain-specific errors
	case ErrTransactionFailed, ErrValidationFailed, ErrCannotCalculateRating:
		return http.StatusUnprocessableEntity

	// Internal server error
	case ErrInternalServer:
		return http.StatusInternalServerError

	// Default internal server error
	default:
		return http.StatusInternalServerError
	}
}
