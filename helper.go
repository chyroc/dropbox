package dropbox

import (
	"encoding/json"
	"fmt"
	"strings"
)

func makeOnlyOnePreSlash(path string) string {
	for strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return "/" + path
}

// null
// {"error_summary": "incorrect_offset/", "error": {".tag": "incorrect_offset", "correct_offset": 0}}
// {"error_summary": "lookup_failed/incorrect_offset/", "error": {".tag": "lookup_failed", "lookup_failed": {".tag": "incorrect_offset", "correct_offset": 1252}}}
func makeDropboxError(bs []byte, typ string) (map[string]interface{}, error) {
	if string(bs) == "null" {
		return nil, nil
	}

	var m = make(map[string]interface{})

	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, NewError(typ, string(bs))
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
		return nil, NewError(typ, errKey)
	}
	return nil, NewError(typ, fmt.Sprintf("%s: %v", errKey, errDetail))
}
