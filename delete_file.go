package dropbox

import (
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) DeleteFile(filename string) (*Metadata, error) {
	url := "https://api.dropboxapi.com/2/files/delete_v2"
	typ := "delete_file"
	body := strings.NewReader(fmt.Sprintf(`{"path":%+q}`, filename))
	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json").WithBody(body)
	var resp struct {
		Metadata *Metadata `json:"metadata"`
	}

	if err := unmarshalResponse(typ, req, &resp); err != nil {
		return nil, err
	}

	return resp.Metadata, nil
}
