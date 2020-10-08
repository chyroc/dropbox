package dropbox

import "io"

type Dropbox interface {
	// file
	FileMetadata(filename string) (*Metadata, error)
	GetFile(filename string) (*Metadata, []byte, error)
	DownloadFile(filename string, dist string) (*Metadata, error)
	DeleteFile(filename string) (*Metadata, error)
	UploadFile(filename string, f io.Reader, overwrite bool) error

	// save url job
	SaveURL(filename, url string) (jobID string, err error)
	CheckSaveURLJob(jobID string) (status string, err error)

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
