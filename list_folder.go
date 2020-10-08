package dropbox

import (
	"encoding/json"
	"net/http"

	"github.com/chyroc/gorequests"
)

func (r *impl) ListFolder(request RequestListFolder) (*ResponseListFolder, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder"

	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json").WithBody(request)
	resp := new(ResponseListFolder)
	if err := unmarshalResponse("list_folder", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type RequestListFolder struct {
	Path                            string `json:"path"`
	Recursive                       bool   `json:"recursive"`
	IncludeMediaInfo                bool   `json:"include_media_info"` // deprecated
	IncludeDeleted                  bool   `json:"include_deleted"`
	IncludeHasExplicitSharedMembers bool   `json:"include_has_explicit_shared_members"`
	IncludeMountedFolders           bool   `json:"include_mounted_folders"`
	IncludeNonDownloadableFiles     bool   `json:"include_non_downloadable_files"`
}

type ResponseListFolder struct {
	Entries []Metadata `json:"entries"`
	Cursor  string     `json:"cursor"`
	HasMore bool       `json:"has_more"`
}

func (r *impl) request(method, url string) *gorequests.Request {
	return gorequests.New(method, url).WithHeader("Authorization", "Bearer "+r.token)
}

func unmarshalResponse(typ string, request *gorequests.Request, resp interface{}) error {
	bs, err := request.Bytes()
	if err != nil {
		return err
	}

	code, err := request.ResponseStatus()
	if err != nil {
		return err
	}

	if code >= 400 {
		_, err = makeDropboxError(bs, typ)
		if err != nil && err.Error() == "not_found" {
			return ErrNotFound
		}
		return err
	}

	if err = json.Unmarshal(bs, resp); err != nil {
		return NewError(typ, err.Error())
	}

	return nil
}
