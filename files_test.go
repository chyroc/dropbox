package dropbox_test

import (
	"strings"
	"testing"

	"github.com/chyroc/dropbox"
)

func TestImpl_UploadFile(t *testing.T) {
	r := dropbox.New("")
	_ = r.UploadFile("1", strings.NewReader("1"), false)
}
