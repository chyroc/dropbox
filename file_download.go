package dropbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) FileDownload(filename string) (*Metadata, []byte, error) {
	url := "https://content.dropboxapi.com/2/files/download"

	headers := map[string]string{
		"Authorization":   "Bearer " + r.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"path":%+q}`, filename),
	}
	resp, bs, err := httpRequest(http.MethodPost, url, nil, headers, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("[dropbox][download] failed: %w", err)
	}

	if resp.Header.Get("Content-Type") == "application/octet-stream" {
		var res = new(Metadata)
		if err := json.Unmarshal([]byte(resp.Header.Get("dropbox-api-result")), res); err != nil {
			return nil, nil, fmt.Errorf("[dropbox][download] 解析结果出错: %+q / %w", bs, err)
		}
		return res, bs, nil
	}

	if _, err = makeDropboxError(bs, "[dropbox][download]"); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, nil, ErrFileNotFound
		}
		return nil, nil, err
	}

	return nil, bs, fmt.Errorf("未知错误")
}
