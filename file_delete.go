package dropbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) FileDelete(filename string) (*Metadata, error) {
	url := "https://api.dropboxapi.com/2/files/delete_v2"

	headers := map[string]string{
		"Authorization": "Bearer " + r.token,
		"Content-Type":  "application/json",
	}
	f := strings.NewReader(fmt.Sprintf(`{"path":%+q}`, filename))
	_, bs, err := httpRequest(http.MethodPost, url, f, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("[dropbox][delete] failed: %w", err)
	}

	if _, err = makeDropboxError(bs, "file_delete"); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var res struct {
		Metadata *Metadata `json:"metadata"`
	}
	if err := json.Unmarshal(bs, &res); err != nil {
		return nil, fmt.Errorf("[dropbox][get metadata] 解析结果出错: %+q / %w", bs, err)
	}

	return res.Metadata, nil
}
