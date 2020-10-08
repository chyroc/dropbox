package dropbox

import "io"

type Dropbox interface {
	// file
	FileMetadata(filename string) (*Metadata, error)
	DownloadFile(filename string) (*Metadata, []byte, error)
	DeleteFile(filename string) (*Metadata, error)
	UploadFile(filename string, f io.Reader, overwrite bool) (err *Error)

	// save url job
	SaveURL(filename, url string) (jobID string, err *Error)
	CheckSaveURLJob(jobID string) (status string, err *Error)

	// list
	ListFolder(request RequestListFolder) (*ResponseListFolder, error)
	ListFolderContinue(cursor string) (*ResponseListFolder, error)
}

type impl struct {
	token string
}

func New(token string) Dropbox {
	return &impl{token: token}
}
