package dropbox

import (
	"fmt"
	"net/http"
	"strings"
)

func (r *impl) SaveURL(filename, fileURL string) (jobID string, err *Error) {
	url := "https://api.dropboxapi.com/2/files/save_url"
	filename = makeOnlyOnePreSlash(filename)

	payload := strings.NewReader(fmt.Sprintf(`{"path":%q,"url":%q}`, filename, fileURL))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", r.token),
		"Content-Type":  "application/json",
	}
	_, bs, err := httpRequest(http.MethodPost, url, payload, headers, nil)
	if err != nil {
		return "", NewError(ErrSaveURLFail, err.Message)
	}

	m, err := makeDropboxError(bs, ErrSaveURLFail)
	if err != nil {
		return "", err
	}

	return m["async_job_id"].(string), nil
}

func (r *impl) CheckSaveURLJob(jobID string) (status string, err *Error) {
	url := "https://api.dropboxapi.com/2/files/save_url/check_job_status"

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", r.token),
		"Content-Type":  "application/json",
	}
	_, bs, err := httpRequest(http.MethodPost, url, strings.NewReader(fmt.Sprintf(`{"async_job_id":%q}`, jobID)), headers, nil)
	if err != nil {
		return "", NewError(ErrGetSaveURLJobFail, err.Message)
	}

	return string(bs), nil
}
