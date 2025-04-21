package domain

import "errors"

type HttpError struct {
	Message string
	Code    int
}

func (e *HttpError) Error() string {
	return e.Message
}

var (
	InternalServerError = &HttpError{"internal server error", 500}
	InvalidEntry        = &HttpError{"Invalid entry", 422}
	UrlNotFound         = &HttpError{"Original url not found", 404}
)

var (
	ErrAlreadyExist = errors.New("row already exist")
)
