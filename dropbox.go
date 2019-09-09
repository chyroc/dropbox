package dropbox

import "io"

type Dropbox interface {
	// file
	UploadFile(filename string, f io.Reader) (err *Error)
	SaveURL(filename, url string) (jobID string, err *Error)
	CheckSaveURLJob(jobID string) (status string, err *Error)
}

type impl struct {
	token string
}

func New(token string) Dropbox {
	return &impl{token: token}
}
