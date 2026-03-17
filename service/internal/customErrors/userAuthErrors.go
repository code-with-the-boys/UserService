package customErrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiError struct {
	Message string     `json:"message"`
	Code    codes.Code `json:"code"`
}

type CustomError struct {
	ApiError apiError `json:"error"`
}

func (e *CustomError) Error() string {
	return e.ApiError.Message
}

func (e *CustomError) GRPCStatus() *status.Status {
	return status.New(e.ApiError.Code, e.ApiError.Message)
}

func NewCustomError(message string, code codes.Code) *CustomError {
	return &CustomError{
		ApiError: apiError{
			Message: message,
			Code:    code,
		},
	}
}

func NewNotFoundError(message string) *CustomError {
	return NewCustomError(message, codes.NotFound)
}

func NewInvalidArgumentError(message string) *CustomError {
	return NewCustomError(message, codes.InvalidArgument)
}

func NewInternalError(message string) *CustomError {
	return NewCustomError(message, codes.Internal)
}

func NewUnauthenticatedError(message string) *CustomError {
	return NewCustomError(message, codes.Unauthenticated)
}

func NewPermissionDeniedError(message string) *CustomError {
	return NewCustomError(message, codes.PermissionDenied)
}

func NewAlreadyExistsError(message string) *CustomError {
	return NewCustomError(message, codes.AlreadyExists)
}

func NewFailedPreconditionError(message string) *CustomError {
	return NewCustomError(message, codes.FailedPrecondition)
}

func NewResourceExhaustedError(message string) *CustomError {
	return NewCustomError(message, codes.ResourceExhausted)
}

func NewUnavailableError(message string) *CustomError {
	return NewCustomError(message, codes.Unavailable)
}

func NewValidationError(message string) *CustomError {
	return NewCustomError(message, codes.InvalidArgument)
}

func NewConflictError(message string) *CustomError {
	return NewCustomError(message, codes.AlreadyExists)
}
