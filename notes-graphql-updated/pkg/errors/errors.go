package errors

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NotFound(msg string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: msg}
}

func Invalid(msg string) *AppError {
	return &AppError{Code: "INVALID_INPUT", Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: "CONFLICT", Message: msg}
}
