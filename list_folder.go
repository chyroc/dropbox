package dropbox

import (
	"encoding/json"
	"net/http"

	"github.com/chyroc/gorequests"
)

// 开始返回文件夹的内容
//
// 如果结果的 has_more 字段为 true，则使用返回的 cursor 调用 list_folder/continue 以检索更多条目
// 如果您使用ListFolderArg.recursive设置为true来保留Dropbox帐户内容的本地缓存，请依次遍历每个条目并按以下方式处理它们，以使本地状态保持同步：
// 对于每个FileMetadata，将新条目存储在本地状态下的给定路径中。如果所需的父文件夹尚不存在，请创建它们。如果给定路径上已经有其他东西，请替换它并删除其所有子元素。
// 对于每个FolderMetadata，将新条目存储在本地状态下的给定路径中。如果所需的父文件夹尚不存在，请创建它们。如果给定的路径上已经有其他东西，请替换它，但让子级保持原样。检查新条目的FolderSharingInfo.read_only并将其所有子级的只读状态设置为匹配。
// 对于每个DeletedMetadata，如果您的本地州在给定的路径上有东西，请将其及其所有子项删除。如果给定路径上没有任何内容，请忽略此条目。
func (r *impl) ListFolder(request RequestListFolder) (*ResponseListFolder, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder"

	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json").WithBody(request)
	resp := new(ResponseListFolder)
	if err := unmarshalResponse("list_folder", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *impl) ListFolderContinue(cursor string) (*ResponseListFolder, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder/continue"

	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json").WithBody(struct {
		Cursor string `json:"cursor"`
	}{Cursor: cursor})
	resp := new(ResponseListFolder)
	if err := unmarshalResponse("list_folder_continue", req, resp); err != nil {
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

func unmarshalBytes(typ string, httpStatusCode int, bs []byte, resp interface{}) error {
	if httpStatusCode >= 400 {
		_, err := makeDropboxError(bs, typ)
		if err != nil && err.Error() == "not_found" {
			return ErrNotFound
		}
		return err
	}

	if resp != nil {
		if err := json.Unmarshal(bs, resp); err != nil {
			return WrapError(typ, err)
		}
	}

	return nil
}

func unmarshalResponse(typ string, request *gorequests.Request, resp interface{}) error {
	bs, err := request.Bytes()
	if err != nil {
		return WrapError(typ, err)
	}

	code, err := request.ResponseStatus()
	if err != nil {
		return WrapError(typ, err)
	}

	return unmarshalBytes(typ, code, bs, resp)
}
