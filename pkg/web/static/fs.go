package static

import (
	"embed"

	"github.com/benbjohnson/hashfs"
)

//go:embed *.css
var fs embed.FS

var HashFS *hashfs.FS

func init() {
	HashFS = hashfs.NewFS(fs)
}

func GetStaticPath(resource string) string {
	return "/static/" + HashFS.HashName(resource)
}
