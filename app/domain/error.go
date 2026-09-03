package domain

import "fmt"

var (
	ErrInvalidArgument = &Error{errType: string(ErrInvalidArgumentType)}
	ErrNotFound        = &Error{errType: string(ErrNotFoundType)}
	ErrAlreadyExists   = &Error{errType: string(ErrAlreadyExistsType)}
	ErrInternal        = &Error{errType: string(ErrInternalType)}
)

type ErrorType string

const (
	ErrInvalidArgumentType ErrorType = "invalid argument"
	ErrNotFoundType        ErrorType = "not found"
	ErrAlreadyExistsType   ErrorType = "already exists"
	ErrInternalType        ErrorType = "internal"
)

type Error struct {
	errType string
	msg     string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf(
			"%s: %s: %v",
			e.errType,
			e.msg,
			e.cause,
		)
	}
	return fmt.Sprintf(
		"%s: %s",
		e.errType,
		e.msg,
	)
}

func (e *Error) With(msg string) *Error {
	return &Error{
		errType: e.errType,
		msg:     msg,
		cause:   e.cause,
	}
}

func (e *Error) Message() string {
	return e.msg
}

func (e *Error) Wrap(cause error, msg string) *Error {
	return &Error{
		errType: e.errType,
		msg:     msg,
		cause:   cause,
	}
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.errType == t.errType
}
