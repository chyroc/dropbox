package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func (r *impl) GetFile(filename string) (*Metadata, []byte, error) {
	url := "https://content.dropboxapi.com/2/files/download"
	typ := "get_file"

	// + -> unicode
	req := r.request(http.MethodPost, url).WithHeader("Dropbox-API-Arg", fmt.Sprintf(`{"path":%+q}`, filename))
	bs, err := req.Bytes()
	if err != nil {
		return nil, nil, err
	}
	headers, err := req.RespHeaders()
	if err != nil {
		return nil, nil, err
	}
	if headers["Content-Type"] == "application/octet-stream" {
		res := new(Metadata)
		if err := json.Unmarshal([]byte(headers["Dropbox-Api-Result"]), res); err != nil {
			return nil, nil, NewError(typ, string(bs))
		}
		return res, bs, nil
	}

	if _, err := makeDropboxError(bs, typ); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	return nil, bs, NewError(typ, "未知错误")
}

func (r *impl) DownloadFile(filename, dist string) (*Metadata, error) {
	url := "https://content.dropboxapi.com/2/files/download"
	typ := "download_file"

	// + -> unicode
	req := r.request(http.MethodPost, url).WithHeader("Dropbox-API-Arg", fmt.Sprintf(`{"path":%+q}`, filename))
	resp, err := req.Response()
	if err != nil {
		return nil, WrapError(typ, err)
	}
	if resp.Header.Get("Content-Type") == "application/octet-stream" {
		meta := new(Metadata)
		if err := json.Unmarshal([]byte(resp.Header.Get("Dropbox-Api-Result")), meta); err != nil {
			return nil, WrapError(typ, err)
		}

		{
			f, err := ioutil.TempFile("", "dropbox-sdk-download-%s")
			if err != nil {
				return nil, WrapError(typ, err)
			}
			defer resp.Body.Close()
			defer f.Close()
			if _, err := io.Copy(f, resp.Body); err != nil {
				return nil, WrapError(typ, err)
			}
			if err := moveFile(f.Name(), dist); err != nil {
				return nil, WrapError(typ, err)
			}
			_ = os.Chtimes(dist, meta.ClientModified, meta.ClientModified)
		}

		return meta, nil
	}

	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapError(typ, err)
	}
	if _, err = makeDropboxError(bs, typ); err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return nil, NewError(typ, "未知错误")
}

func moveFile(source, dist string) error {
	distDir := dist[:strings.LastIndex(dist, "/")]
	err := os.MkdirAll(distDir, 0o777)
	if err != nil {
		return err
	}
	cmd := exec.Command("mv", source, dist)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
