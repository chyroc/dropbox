package dropbox

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not_found")

type SharingInfo struct {
	ReadOnly             bool   `json:"read_only"`
	ParentSharedFolderID string `json:"parent_shared_folder_id"`
	ModifiedBy           string `json:"modified_by"`
}

type FileLockInfo struct {
	IsLockHolder        bool      `json:"is_lockholder"`
	LockHolderName      string    `json:"lockholder_name"`
	LockHolderAccountID string    `json:"lockholder_account_id"`
	Created             time.Time `json:"created"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PropertyGroup struct {
	TemplateID string  `json:"template_id"`
	Fields     []Field `json:"fields"`
}

type SymlinkInfo struct {
	Target string `json:"target"`
}

type ExportInfo struct {
	ExportAs string `json:"export_as"`
}

type Tag string

const (
	TagFile   Tag = "file"
	TagFolder Tag = "folder"
)

type Metadata struct {
	Tag                      Tag             `json:".tag"`
	Name                     string          `json:"name"`
	ID                       string          `json:"id"`
	ClientModified           time.Time       `json:"client_modified"`
	ServerModified           time.Time       `json:"server_modified"`
	Rev                      string          `json:"rev"`
	Size                     int             `json:"size"`
	PathLower                string          `json:"path_lower"`
	PathDisplay              string          `json:"path_display"`
	SymlinkInfo              SymlinkInfo     `json:"symlink_info"`
	SharingInfo              SharingInfo     `json:"sharing_info"`
	IsDownloadable           bool            `json:"is_downloadable"`
	ExportInfo               ExportInfo      `json:"export_info"`
	PropertyGroups           []PropertyGroup `json:"property_groups"`
	HasExplicitSharedMembers bool            `json:"has_explicit_shared_members"`
	ContentHash              string          `json:"content_hash"`
	FileLockInfo             FileLockInfo    `json:"file_lock_info"`
}

func (r *impl) FileMetadata(filename string) (*Metadata, error) {
	url := "https://api.dropboxapi.com/2/files/get_metadata"
	typ := "file_metadata"
	req := r.request(http.MethodPost, url).WithHeader("Content-Type", "application/json").WithBody(strings.NewReader(fmt.Sprintf(`{"path":%+q,"include_media_info":true}`, filename)))
	resp := new(Metadata)

	if err := unmarshalResponse(typ, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
