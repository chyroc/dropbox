package dropbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) DownloadFile(filename string) (*Metadata, []byte, error) {
	url := "https://content.dropboxapi.com/2/files/download"
	typ := "download_file"

	// + -> unicode
	req := r.request(http.MethodPost, url).WithHeader("Dropbox-API-Arg", fmt.Sprintf(`{"path":%+q}`, filename))
	bs, err := req.Bytes()
	if err != nil {
		return nil, nil, err
	}
	headers, err := req.RespHeaders()
	if err != nil {
		return nil, nil, err
	}
	if headers["Content-Type"] == "application/octet-stream" {
		var res = new(Metadata)
		if err := json.Unmarshal([]byte(headers["Dropbox-Api-Result"]), res); err != nil {
			return nil, nil, NewError(typ, string(bs))
		}
		return res, bs, nil
	}

	if _, err = makeDropboxError(bs, typ); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	return nil, bs, NewError(typ, "未知错误")
}
