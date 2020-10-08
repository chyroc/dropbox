package dropbox

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func (r *impl) UploadFile(filename string, f io.Reader, overwrite bool) error {
	// defer printTrace()

	filename = makeOnlyOnePreSlash(filename)

	var buf = make([]byte, MaxSingleUploadFileSize) // 150M
	readLen, err2 := f.Read(buf)
	if err2 != nil && err2 != io.EOF {
		return NewError("upload_file", err2.Error())
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
			return NewError("upload_file", err2.Error())
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

func (r *impl) uploadFile(filename string, f io.Reader, overwrite bool) (err error) {
	url := "https://content.dropboxapi.com/2/files/upload"
	typ := "upload_file"

	mode := ""
	if overwrite {
		mode = "overwrite"
	} else {
		mode = "add"
	}

	headers := map[string]string{
		"Dropbox-API-Arg": fmt.Sprintf(`{"path": %+q,"mode": %+q,"autorename": true,"mute": false,"strict_conflict": false}`, filename, mode),
		"Content-Type":    "application/octet-stream",
	}
	req := r.request(http.MethodPost, url).WithHeaders(headers).WithBody(f)
	return unmarshalResponse(typ, req, nil)
}

type uploadSession struct {
	sessionID string
	offset    int
	token     string
	dropbox   *impl
}

func (r *impl) startSession(f io.Reader, length int) (*uploadSession, error) {
	url := "https://content.dropboxapi.com/2/files/upload_session/start"
	typ := "upload_file_start"
	headers := map[string]string{
		"Dropbox-API-Arg": `{"close":false}`,
		"Content-Type":    "application/octet-stream",
	}
	req := r.request(http.MethodPost, url).WithHeaders(headers).WithBody(f)
	resp := make(map[string]interface{})
	if err := unmarshalResponse(typ, req, &resp); err != nil {
		return nil, err
	}

	return &uploadSession{sessionID: resp["session_id"].(string), offset: length, token: r.token, dropbox: r}, nil
}

func (s *uploadSession) appendSession(f io.Reader, length int) error {
	url := "https://content.dropboxapi.com/2/files/upload_session/append_v2"
	typ := "upload_file_append"
	headers := map[string]string{
		"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%+q,"offset":%d},"close":false}`, s.sessionID, s.offset),
		"Content-Type":    "application/octet-stream",
	}
	req := s.dropbox.request(http.MethodPost, url).WithHeaders(headers).WithBody(f)
	resp := make(map[string]interface{})
	if err := unmarshalResponse(typ, req, &resp); err != nil {
		return err
	}

	s.offset += length
	return nil
}

func (s *uploadSession) finishSession(filename string, overwrite bool) error {
	url := "https://content.dropboxapi.com/2/files/upload_session/finish"
	typ := "upload_file_finish"
	headers := map[string]string{}
	{
		mode := ""
		if overwrite {
			mode = "overwrite"
		} else {
			mode = "add"
		}
		headers = map[string]string{
			"Dropbox-API-Arg": fmt.Sprintf(`{"cursor":{"session_id":%+q,"offset":%d},"commit":{"path":%+q,"mode":%+q,"autorename":true,"mute":false,"strict_conflict":false}}`, s.sessionID, s.offset, filename, mode),
			"Content-Type":    "application/octet-stream",
		}
	}
	req := s.dropbox.request(http.MethodPost, url).WithHeaders(headers)
	resp := make(map[string]interface{})

	if err := unmarshalResponse(typ, req, &resp); err != nil {
		return err
	}

	return nil
}
