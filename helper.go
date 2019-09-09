package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func httpRequest(method, url string, body io.Reader, headers map[string]string, res interface{}) (int, []byte, *Error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, NewError(ErrRequestFail, fmt.Sprintf("%s; method=%s, url=%s", err, method, url))
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, NewError(ErrRequestFail, err.Error())
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, NewError(ErrRequestFail, err.Error())
	}

	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			return 0, bs, NewError(ErrRequestFail, err.Error())
		}
	}

	return resp.StatusCode, bs, nil
}
