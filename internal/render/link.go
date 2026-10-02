package render

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// localHost returns the host of file URLs: the name of this machine, so a
// terminal can tell local files from remote ones. It is empty when the name is
// unknown and on Windows, where the host of a file URL names a network share.
var localHost = sync.OnceValue(func() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	host, _ := os.Hostname()
	return host
})

// destination returns the URL of a link or an image. With hyperlinks, a
// relative path is resolved against the directory of the document; in plain
// mode the destination is shown as written, which reads better.
func (r *renderer) destination(dest []byte) string {
	if !r.opts.Style.Hyperlinks || r.opts.Dir == "" {
		return string(dest)
	}
	return resolveLink(string(dest), r.opts.Dir, localHost())
}

// resolveLink returns the URL of the link destination dest of a document in
// the directory dir: a relative path becomes a file URL on host, anything
// else (a scheme, a fragment only, a leading slash) is returned as written.
func resolveLink(dest, dir, host string) string {
	if dir == "" || strings.HasPrefix(dest, "/") {
		return dest
	}
	u, err := url.Parse(string(unescape([]byte(dest))))
	if err != nil || u.Scheme != "" || u.Path == "" {
		return dest
	}
	return fileURL(host, filepath.Join(dir, filepath.FromSlash(u.Path)), u.Fragment)
}

// fileURL returns the file URL of the absolute path p on host. The query of a
// destination means nothing for a file, so only the fragment is kept.
func fileURL(host, p, fragment string) string {
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "/") {
		// A Windows drive path: C:/proj becomes /C:/proj.
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Host: host, Path: p, Fragment: fragment}
	return u.String()
}
