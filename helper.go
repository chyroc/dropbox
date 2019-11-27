package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
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
	fmt.Println("httpRequest.url", url)
	fmt.Println("httpRequest.headers", headers)
	fmt.Println("httpRequest.res", string(bs))

	return resp.StatusCode, bs, nil
}

func printTrace() {
	if err := recover(); err != nil {
		pc, _, _, _ := runtime.Caller(3)
		f := runtime.FuncForPC(pc)
		fmt.Printf("functipn=%v\npanic:%v\nstack info:%v", f.Name(), err, string(debug.Stack()))
	}
}

func makeOnlyOnePreSlash(path string) string {
	for strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return "/" + path
}

// null
// {"error_summary": "incorrect_offset/", "error": {".tag": "incorrect_offset", "correct_offset": 0}}
// {"error_summary": "lookup_failed/incorrect_offset/", "error": {".tag": "lookup_failed", "lookup_failed": {".tag": "incorrect_offset", "correct_offset": 1252}}}
func makeDropboxError(bs []byte, msg string) (map[string]interface{}, *Error) {
	if string(bs) == "null" {
		return nil, nil
	}

	var m = make(map[string]interface{})

	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, NewError(msg, fmt.Sprintf("decode json fail: %s", err))
	}
	if _, ok := m["error"]; !ok {
		return m, nil
	}

	var errDetail interface{}
	var errKey string

	var getErrDetail = func(er map[string]interface{}, key string) string {
		if _, ok := er[".tag"]; ok {
			delete(er, ".tag")
		}
		b, _ := json.Marshal(er)
		return string(b)
	}

	summary := m["error_summary"].(string)
	summarys := strings.Split(summary, "/")
	errMap := m["error"].(map[string]interface{})
	for i := 0; i < len(summarys)-1; i++ {
		if i == len(summarys)-2 {
			errDetail = getErrDetail(errMap, summarys[i])
			errKey = summarys[i]
			break
		}

		errMap = errMap[summarys[i]].(map[string]interface{})
		continue
	}

	if errDetail == "{}" {
		return nil, NewError(msg, errKey)
	}
	return nil, NewError(msg, fmt.Sprintf("%s: %v", errKey, errDetail))
}
