package dropbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrFileNotFound = errors.New("file not found")

type Metadata struct {
	Name           string    `json:"name"`
	PathLower      string    `json:"path_lower"`
	PathDisplay    string    `json:"path_display"`
	ID             string    `json:"id"`
	ClientModified time.Time `json:"client_modified"`
	ServerModified time.Time `json:"server_modified"`
	Rev            string    `json:"rev"`
	Size           int       `json:"size"`
	IsDownloadable bool      `json:"is_downloadable"`
	ContentHash    string    `json:"content_hash"`
}

func (r *impl) FileMetadata(filename string) (*Metadata, error) {
	url := "https://api.dropboxapi.com/2/files/get_metadata"

	headers := map[string]string{
		"Authorization": "Bearer " + r.token,
		"Content-Type":  "application/json",
	}
	f := strings.NewReader(fmt.Sprintf(`{"path":%+q,"include_media_info":true}`, filename))

	_, bs, err := httpRequest(http.MethodPost, url, f, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("[dropbox][get metadata] failed: %w", err)
	}

	if _, err = makeDropboxError(bs, "[dropbox][get metadata]"); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	var res = new(Metadata)
	if err := json.Unmarshal(bs, res); err != nil {
		return nil, fmt.Errorf("[dropbox][get metadata] 解析结果出错: %+q / %w", bs, err)
	}
	return res, nil
}
