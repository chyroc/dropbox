package dropbox

import "io"

type Dropbox interface {
	// file
	FileMetadata(filename string) (*Metadata, error)
	FileDownload(filename string) (*Metadata, []byte, error)
	FileDelete(filename string) (*Metadata, error)
	UploadFile(filename string, f io.Reader, overwrite bool) (err *Error)
	SaveURL(filename, url string) (jobID string, err *Error)
	CheckSaveURLJob(jobID string) (status string, err *Error)

	// files
	ListFolder(request RequestListFolder) (*ResponseListFolder, error)
}

type impl struct {
	token string
}

func New(token string) Dropbox {
	return &impl{token: token}
}
