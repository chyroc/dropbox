package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	return nil, NewError(msg, fmt.Sprintf("%s: %v", errKey, errDetail))
}

func (r *impl) UploadFile(filename string, f io.Reader, overwrite bool) (err *Error) {
	defer printTrace()

	filename = makeOnlyOnePreSlash(filename)

	var buf = make([]byte, MaxSingleUploadFileSize) // 150M
	readLen, err2 := f.Read(buf)
	if err2 != nil && err2 != io.EOF {
		return NewError(ErrReadFileFail, err2.Error())
	}
	if readLen < MaxSingleUploadFileSize {
		// 这里小于 150 M，那么就直接调用 upload 接口就行
		log.Printf("[dropbox][UploadFile] small file\n")
		return r.uploadFile(filename, bytes.NewReader(buf[:readLen]), overwrite)
	}

	// 否则分片
	log.Printf("[dropbox][UploadFile] big file\n")
	session, err := r.startSession(bytes.NewReader(buf[:readLen]), readLen)
	if err != nil {
		return err
	}
	for {
		readLen, err2 := f.Read(buf)
		if err2 != nil && err2 != io.EOF {
			return NewError(ErrReadFileFail, err2.Error())
		}
		if readLen == 0 {
			return session.finishSession(filename, overwrite)
		}

		log.Printf("[dropbox][UploadFile] use append api\n")
		if err := session.appendSession(bytes.NewReader(buf[:readLen]), readLen); err != nil {
			return err
		}
	}
}

func (r *impl) uploadFile(filename string, f io.Reader, overwrite bool) (err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload"

	mode := ""
	if overwrite {
		mode = "overwrite"
	} else {
		mode = "add"
	}

	headers := map[string]string{
		"Authorization":   "Bearer " + r.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"path": %+q,"mode": %+q,"autorename": true,"mute": false,"strict_conflict": false}`, filename, mode),
		"Content-Type":    "application/octet-stream",
	}
	fmt.Println("headers", headers)

	if _, _, err := httpRequest(http.MethodPost, url, f, headers, nil); err != nil {
		return err
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

	_, bs, err := httpRequest(http.MethodPost, url, f, headers, nil)
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
		"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%+q,"offset":%d},"close":false}`, s.sessionID, s.offset),
		"Content-Type":    "application/octet-stream",
	}

	_, bs, err := httpRequest(http.MethodPost, url, f, headers, nil)
	if err != nil {
		return NewError(ErrUploadFileAppendFail, err.Message)
	}

	_, err = makeDropboxError(bs, ErrUploadFileAppendFail)
	if err != nil {
		return err
	}
	s.offset += length
	return nil
}

func (s *uploadSession) finishSession(filename string, overwrite bool) (err *Error) {
	url := "https://content.dropboxapi.com/2/files/upload_session/finish"

	mode := ""
	if overwrite {
		mode = "overwrite"
	} else {
		mode = "add"
	}
	headers := map[string]string{
		"Authorization":   "Bearer " + s.token,
		"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%+q,"offset":%d},"commit":{"path":%+q,"mode":%+q,"autorename":true,"mute":false,"strict_conflict":false}}`, s.sessionID, s.offset, filename, mode),
		"Content-Type":    "application/octet-stream",
	}

	_, bs, err := httpRequest(http.MethodPost, url, nil, headers, nil)
	if err != nil {
		return NewError(ErrUploadFileFinishFail, err.Message)
	}

	_, err = makeDropboxError(bs, ErrUploadFileFinishFail)
	return err
}
