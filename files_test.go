package dropbox_test

import (
	"github.com/chyroc/dropbox"
	"strings"
	"testing"
)

func TestImpl_UploadFile(t *testing.T) {
	r := dropbox.New("")
	_ = r.UploadFile("1", strings.NewReader("1"), false)
}
