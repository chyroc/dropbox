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

const (
	ErrReadFileFail         = "read_file_fail"
	ErrRequestFail          = "request_fail"
	ErrUploadFileStartFail  = "upload_start_fail"
	ErrUploadFileAppendFail = "upload_append_fail"
	ErrUploadFileFinishFail = "upload_finish_fail"
)
