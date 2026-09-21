package errors

import (
	stderrors "errors"
)

// Is は stdlib errors.Is を呼び出す。
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As は stdlib errors.As を呼び出す。
func As(err error, target any) bool { return stderrors.As(err, target) }

// New は stdlib errors.New を呼び出す。
func New(text string) error { return stderrors.New(text) }
