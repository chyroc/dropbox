package dropbox

import "fmt"

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[dropbox] %s: %s", e.Type, e.Message)
}

func NewError(typ, msg string) *Error {
	return &Error{
		Type:    typ,
		Message: msg,
	}
}
