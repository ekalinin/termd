package render

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/ekalinin/termd/internal/text"
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

// linkRefs holds the URLs of the reference markers of a table, in the order
// of their numbers.
type linkRefs struct {
	urls []string
}

// marker returns the number of the marker of dest, adding dest when it is
// new.
func (l *linkRefs) marker(dest string) int {
	if i := slices.Index(l.urls, dest); i >= 0 {
		return i + 1
	}
	l.urls = append(l.urls, dest)
	return len(l.urls)
}

// link turns the label of a link to dest into spans. In a table cell without
// hyperlinks, the label is followed by a reference marker instead of the URL,
// and a bare link, which is bare or has a label that reads as dest, shows a
// short text instead of the URL; elsewhere the spans are those of
// text.LinkSpans.
func (r *renderer) link(label []text.Span, dest string, bare bool) []text.Span {
	if r.refs == nil || dest == "" {
		return text.LinkSpans(label, dest, r.opts.Style.Hyperlinks)
	}
	if bare || text.Line(label).String() == dest {
		short, ok := shortText(dest)
		if !ok {
			return label
		}
		label = []text.Span{{Text: short, Style: label[0].Style}}
	}
	return append(slices.Clip(label), text.Span{Text: " [" + strconv.Itoa(r.refs.marker(dest)) + "]"})
}

// shortText returns the text of a bare link to dest in a table cell: the
// GitHub text of the URL, or else its host. It returns false for a URL
// without a host.
func shortText(dest string) (string, bool) {
	if s, ok := githubText(dest); ok {
		return s, true
	}
	u, err := url.Parse(dest)
	if err != nil || u.Host == "" {
		return "", false
	}
	return u.Host, true
}

// htmlDestination returns the URL of an href or src attribute, which the
// tokenizer has already decoded, under the rules of destination.
func (r *renderer) htmlDestination(v string) string {
	if !r.opts.Style.Hyperlinks || r.opts.Dir == "" {
		return v
	}
	return resolveDecoded(v, v, r.opts.Dir, localHost())
}

// resolveLink returns the URL of the link destination dest of a document in
// the directory dir: a relative path becomes a file URL on host, anything
// else (a scheme, a fragment only, a leading slash) is returned as written.
func resolveLink(dest, dir, host string) string {
	return resolveDecoded(dest, string(unescape([]byte(dest))), dir, host)
}

// resolveDecoded is resolveLink for a destination written as dest, whose
// escapes and character references decode to decoded.
func resolveDecoded(dest, decoded, dir, host string) string {
	if dir == "" || strings.HasPrefix(dest, "/") {
		return dest
	}
	u, err := url.Parse(decoded)
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
