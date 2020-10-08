package dropbox

import (
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) SaveURL(filename, fileURL string) (jobID string, err error) {
	url := "https://api.dropboxapi.com/2/files/save_url"
	typ := "save_url"
	body := strings.NewReader(fmt.Sprintf(`{"path":%+q,"url":%+q}`, makeOnlyOnePreSlash(filename), fileURL))
	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json; charset=utf-8").WithBody(body)
	resp := make(map[string]interface{})

	if err = unmarshalResponse(typ, req, &resp); err != nil {
		return "", err
	}

	return resp["async_job_id"].(string), nil
}

func (r *impl) CheckSaveURLJob(jobID string) (status string, err error) {
	url := "https://api.dropboxapi.com/2/files/save_url/check_job_status"
	typ := "check_save_url_job"
	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json; charset=utf-8").WithBody(strings.NewReader(fmt.Sprintf(`{"async_job_id":%+q}`, jobID)))
	resp := make(map[string]interface{})

	if err = unmarshalResponse(typ, req, &resp); err != nil {
		return "", err
	}

	return resp[".tag"].(string), nil
}
