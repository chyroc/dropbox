package dropbox

import (
	"errors"
	"fmt"
)

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	err     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("[dropbox] %s: %s", e.Type, e.Message)
}

func (e *Error) Is(e2 error) bool {
	if e == nil && e2 == nil {
		return true
	}
	if (e == nil && e2 != nil) || (e != nil && e2 == nil) {
		return false
	}

	e3, ok := e2.(*Error)
	if ok && e3.Type == e.Type && e3.Message == e.Message {
		return true
	}

	return errors.Is(e.err, e2)
}

func (e *Error) Unwrap() error {
	return e.err
}

func NewError(typ, msg string) error {
	return &Error{
		Type:    typ,
		Message: msg,
	}
}

func WrapError(typ string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Type:    typ,
		Message: err.Error(),
		err:     err,
	}
}
