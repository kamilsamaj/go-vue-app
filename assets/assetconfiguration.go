// +build dev


package assets

import (
	"net/http"
	"os"
	pathpkg "path"

	"github.com/shurcooL/httpfs/filter"
)

/*
 * Skip files we don't need to include
 */
var skipFunc = func(path string, fi os.FileInfo) bool {
	fname := fi.Name()
	extension := pathpkg.Ext(fname)

	return extension == ".go" ||
		extension == ".DS_Store" ||
		extension == ".md" ||
		extension == ".svg" ||
		fname == "LICENSE"
}

var Assets = filter.Skip(http.Dir("./app/dist/"), skipFunc)
