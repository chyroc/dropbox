package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	summary := m["error_summary"].(string)
	summarys := strings.Split(summary, "/")
	errMap := m["error"].(map[string]interface{})
	for i := 0; i < len(summarys)-1; i++ {
		if i == len(summarys)-2 {
			errDetail = errMap[summarys[i]]
			errKey = summarys[i]
		}

		errMap = errMap[summarys[i]].(map[string]interface{})
		continue
	}

	return nil, NewError(msg, fmt.Sprintf("%s: %v", errKey, errDetail))
}

func (r *impl) UploadFile(filename string, f io.Reader) (err *Error) {
	filename = makeOnlyOnePreSlash(filename)

	var buf = make([]byte, MaxSingleUploadFileSize) // 150M
	readLen, err2 := f.Read(buf)
	if err2 != nil {
		return NewError(ErrReadFileFail, err.Error())
	}
	if readLen < MaxSingleUploadFileSize {
		// 这里小于 150 M，那么就直接调用 upload 接口就行
		return r.uploadFile(filename, bytes.NewReader(buf[:readLen]))
	}

	// 否则分片
	session, err := r.startSession(bytes.NewReader(buf[:readLen]), readLen)
	if err != nil {
		return err
	}
	for {
		readLen, err2 := f.Read(buf)
		if err2 != nil {
			return NewError(ErrReadFileFail, err.Error())
		}
		if readLen == 0 {
			return session.finishSession(filename)
		}

		if err = session.appendSession(bytes.NewReader(buf[:readLen]), readLen); err != nil {
			return err
		}
	}
}

func (r *impl) uploadFile(filename string, f io.Reader) (err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload"

	headers := map[string]string{
		"Authorization":   "Bearer " + r.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"path": %q,"mode": "add","autorename": true,"mute": false,"strict_conflict": false}`, filename),
		"Content-Type":    "application/octet-stream",
	}

	if _, bs, err := httpRequest(url, http.MethodPost, f, headers, nil); err != nil {
		return err
	} else {
		fmt.Println(string(bs))
	}

	return nil
}

type uploadSession struct {
	sessionID string
	offset    int
	token     string
}

func (r *impl) startSession(f io.Reader, length int) (session *uploadSession, err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload_session/start"

	headers := map[string]string{
		"Authorization":   "Bearer " + r.token,
		"Dropbox-API-Arg": `{"close":false}`,
		"Content-Type":    "application/octet-stream",
	}

	_, bs, err := httpRequest(url, http.MethodPost, f, headers, nil)
	if err != nil {
		return nil, NewError(ErrUploadFileStartFail, err.Message)
	}

	m, err := makeDropboxError(bs, ErrUploadFileStartFail)
	if err != nil {
		return nil, err
	}

	return &uploadSession{sessionID: m["session_id"].(string), offset: length, token: r.token}, nil
}

func (s *uploadSession) appendSession(f io.Reader, length int) (err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload_session/append_v2"

	headers := map[string]string{
		"Authorization":   "Bearer " + s.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%q,"offset":%d},"close":false}`, s.sessionID, s.offset),
		"Content-Type":    "application/octet-stream",
	}

	_, bs, err := httpRequest(url, http.MethodPost, f, headers, nil);
	if err != nil {
		return NewError(ErrUploadFileAppendFail, err.Message)
	}

	_, err = makeDropboxError(bs, ErrUploadFileAppendFail)
	return err
}

func (s *uploadSession) finishSession(filename string) (err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload_session/finish"

	headers := map[string]string{
		"Authorization":   "Bearer " + s.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%q,"offset":%d},"commit":{"path":%q,"mode":"add","autorename":true,"mute":false,"strict_conflict":false}}`, s.sessionID, s.offset, filename),
		"Content-Type":    "application/octet-stream",
	}

	_, bs, err := httpRequest(url, http.MethodPost, nil, headers, nil)
	if err != nil {
		return NewError(ErrUploadFileFinishFail, err.Message)
	}

	_, err = makeDropboxError(bs, ErrUploadFileFinishFail)
	return err
}
