package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rivo/uniseg"
)

// fake is a scripted environment for run.
type fake struct {
	stdin       string
	stdinTTY    bool
	stdoutTTY   bool
	width       int
	height      int
	colorterm   string
	hasLess     bool
	light       bool
	lightOK     bool
	stdout      bytes.Buffer
	stderr      bytes.Buffer
	pagerCalls  []string
	detectCalls int
}

func (f *fake) env() env {
	return env{
		stdin:     strings.NewReader(f.stdin),
		stdout:    &f.stdout,
		stderr:    &f.stderr,
		stdinTTY:  f.stdinTTY,
		stdoutTTY: f.stdoutTTY,
		size:      func() (int, int, bool) { return f.width, f.height, f.width > 0 },
		getenv: func(k string) string {
			if k == "COLORTERM" {
				return f.colorterm
			}
			return ""
		},
		readFile: os.ReadFile,
		stat:     os.Stat,
		readDir:  os.ReadDir,
		lookPath: func(name string) (string, error) {
			if f.hasLess && name == "less" {
				return "/usr/bin/less", nil
			}
			return "", errors.New("not found")
		},
		runPager: func(path, content string) error {
			f.pagerCalls = append(f.pagerCalls, content)
			return nil
		},
		detectLight: func() (bool, bool) {
			f.detectCalls++
			return f.light, f.lightOK
		},
	}
}

func (f *fake) run(args ...string) int {
	return run(args, f.env())
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	return writeNamedFile(t, "doc.md", content)
}

func writeNamedFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeDir creates a temp directory with the given entries: a name ending in
// / is a directory, any other name a file holding the heading "# <name>".
func writeDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

var longParagraph = strings.Repeat("words that wrap ", 40) + "\n"

func maxLineWidth(s string) int {
	w := 0
	for l := range strings.SplitSeq(s, "\n") {
		w = max(w, uniseg.StringWidth(l))
	}
	return w
}

func TestInvalidFlags(t *testing.T) {
	for _, args := range [][]string{
		{"--width", "0"},
		{"--width", "-5"},
		{"--width", "abc"},
		{"--hyperlinks=sometimes"},
		{"--theme=blue"},
		{"a.md", "b.md"},
	} {
		f := &fake{stdin: "# x\n"}
		if code := f.run(args...); code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
		if !strings.Contains(f.stderr.String(), "Usage: termd") {
			t.Errorf("%v: no usage on stderr: %q", args, f.stderr.String())
		}
	}
}

func TestInputSources(t *testing.T) {
	path := writeFile(t, "# From file\n")

	t.Run("file", func(t *testing.T) {
		f := &fake{stdinTTY: true}
		if code := f.run(path); code != 0 || !strings.Contains(f.stdout.String(), "# From file") {
			t.Errorf("exit %d, stdout %q", code, f.stdout.String())
		}
	})
	t.Run("piped stdin", func(t *testing.T) {
		f := &fake{stdin: "# From stdin\n"}
		if code := f.run(); code != 0 || !strings.Contains(f.stdout.String(), "# From stdin") {
			t.Errorf("exit %d, stdout %q", code, f.stdout.String())
		}
	})
	t.Run("explicit stdin", func(t *testing.T) {
		f := &fake{stdin: "# Dash\n", stdinTTY: true}
		if code := f.run("-"); code != 0 || !strings.Contains(f.stdout.String(), "# Dash") {
			t.Errorf("exit %d, stdout %q", code, f.stdout.String())
		}
	})
	t.Run("no input", func(t *testing.T) {
		f := &fake{stdinTTY: true}
		if code := f.run(); code != 2 || !strings.Contains(f.stderr.String(), "Usage: termd") {
			t.Errorf("exit %d, stderr %q", code, f.stderr.String())
		}
	})
	t.Run("unreadable file", func(t *testing.T) {
		f := &fake{stdinTTY: true}
		if code := f.run("missing.md"); code != 1 || !strings.Contains(f.stderr.String(), "missing.md") {
			t.Errorf("exit %d, stderr %q", code, f.stderr.String())
		}
	})
}

func TestDirectoryInput(t *testing.T) {
	for _, tt := range []struct {
		name  string
		files []string
		want  string
	}{
		{"README next to other files", []string{"README.md", "guide.md"}, "README.md"},
		{"README.md before the others", []string{"README", "README.markdown", "README.md"}, "README.md"},
		{"README.markdown before README", []string{"README", "README.markdown"}, "README.markdown"},
		{"lowercase name", []string{"readme.md"}, "readme.md"},
		{"directory named README.md", []string{"README.md/", "README"}, "README"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := &fake{stdinTTY: true}
			code := f.run(writeDir(t, tt.files...))
			if want := "# " + tt.want + "\n"; code != 0 || f.stdout.String() != want {
				t.Errorf("exit %d, stdout %q, want %q; stderr %q", code, f.stdout.String(), want, f.stderr.String())
			}
		})
	}
	for _, tt := range []struct {
		name  string
		files []string
	}{
		{"README only in a subdirectory", []string{"guide/README.md"}},
		{"no README", []string{"guide.md"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeDir(t, tt.files...) + string(filepath.Separator)
			f := &fake{stdinTTY: true}
			code := f.run(dir)
			if want := "termd: no README in " + dir + "\n"; code != 1 || f.stdout.Len() != 0 || f.stderr.String() != want {
				t.Errorf("exit %d, stdout %q, stderr %q, want %q", code, f.stdout.String(), f.stderr.String(), want)
			}
		})
	}
}

func TestDirectorySymlinks(t *testing.T) {
	symlink := func(t *testing.T, target, link string) {
		t.Helper()
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("cannot create a symbolic link: %v", err)
		}
	}
	t.Run("link to a directory", func(t *testing.T) {
		dir := writeDir(t, "docs/README.md")
		link := filepath.Join(dir, "link")
		symlink(t, "docs", link)
		f := &fake{stdinTTY: true}
		if code := f.run(link); code != 0 || f.stdout.String() != "# docs/README.md\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, f.stdout.String(), f.stderr.String())
		}
	})
	t.Run("README linked to a file", func(t *testing.T) {
		dir := writeDir(t, "README.md", "docs/")
		symlink(t, filepath.Join("..", "README.md"), filepath.Join(dir, "docs", "README.md"))
		f := &fake{stdinTTY: true}
		if code := f.run(filepath.Join(dir, "docs")); code != 0 || f.stdout.String() != "# README.md\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, f.stdout.String(), f.stderr.String())
		}
	})
	t.Run("broken README link is skipped", func(t *testing.T) {
		dir := writeDir(t, "README")
		symlink(t, "missing.md", filepath.Join(dir, "README.md"))
		f := &fake{stdinTTY: true}
		if code := f.run(dir); code != 0 || f.stdout.String() != "# README\n" {
			t.Errorf("exit %d, stdout %q, stderr %q", code, f.stdout.String(), f.stderr.String())
		}
	})
}

func TestDirectoryCaseVariants(t *testing.T) {
	// A case-insensitive file system cannot hold both names, so the
	// directory is faked.
	m := fstest.MapFS{
		"docs/readme.md": {Data: []byte("# lower\n")},
		"docs/README.md": {Data: []byte("# upper\n")},
	}
	f := &fake{stdinTTY: true}
	e := f.env()
	e.stat = func(name string) (os.FileInfo, error) { return fs.Stat(m, filepath.ToSlash(name)) }
	e.readDir = func(name string) ([]os.DirEntry, error) { return fs.ReadDir(m, filepath.ToSlash(name)) }
	e.readFile = func(name string) ([]byte, error) { return fs.ReadFile(m, filepath.ToSlash(name)) }
	if code := run([]string{"docs"}, e); code != 0 || f.stdout.String() != "# upper\n" {
		t.Errorf("exit %d, stdout %q, stderr %q", code, f.stdout.String(), f.stderr.String())
	}
}

func TestDirectoryReadErrors(t *testing.T) {
	t.Run("unreadable README", func(t *testing.T) {
		dir := writeDir(t, "README.md", "README")
		readme := filepath.Join(dir, "README.md")
		f := &fake{stdinTTY: true}
		e := f.env()
		e.readFile = func(name string) ([]byte, error) {
			if name == readme {
				return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrPermission}
			}
			return os.ReadFile(name)
		}
		if code := run([]string{dir}, e); code != 1 || f.stdout.Len() != 0 || !strings.Contains(f.stderr.String(), readme) {
			t.Errorf("exit %d, stdout %q, stderr %q", code, f.stdout.String(), f.stderr.String())
		}
	})
	t.Run("directory cannot be listed", func(t *testing.T) {
		dir := writeDir(t, "README.md")
		f := &fake{stdinTTY: true}
		e := f.env()
		e.readDir = func(name string) ([]os.DirEntry, error) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrPermission}
		}
		if code, want := run([]string{dir}, e), "termd: open "+dir+": permission denied\n"; code != 1 || f.stderr.String() != want {
			t.Errorf("exit %d, stderr %q, want %q", code, f.stderr.String(), want)
		}
	})
}

func TestDirectoryRendersAsFile(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte(strings.Repeat("**bold** "+longParagraph+"\n", 5)), 0o644); err != nil {
		t.Fatal(err)
	}
	var pages []string
	for _, arg := range []string{dir, readme} {
		f := &fake{stdoutTTY: true, width: 80, height: 24, hasLess: true}
		if code := f.run("--width", "60", arg); code != 0 || len(f.pagerCalls) != 1 || f.stdout.Len() != 0 {
			t.Fatalf("%s: exit %d, pager calls %d, stdout %q, stderr %q", arg, code, len(f.pagerCalls), f.stdout.String(), f.stderr.String())
		}
		pages = append(pages, f.pagerCalls[0])
	}
	if pages[0] != pages[1] {
		t.Errorf("directory:\n%s\nREADME:\n%s", pages[0], pages[1])
	}
}

func TestRelativeLinks(t *testing.T) {
	const doc = "[guide](docs/guide.md)\n"
	path := writeFile(t, doc)
	dir := filepath.Dir(path)
	host, _ := os.Hostname()
	resolved := "\x1b]8;;file://" + host + filepath.ToSlash(dir) + "/docs/guide.md\x1b\\guide"
	asWritten := "\x1b]8;;docs/guide.md\x1b\\guide"

	t.Run("absolute path", func(t *testing.T) {
		f := &fake{stdinTTY: true}
		if f.run("--hyperlinks=always", path); !strings.Contains(f.stdout.String(), resolved) {
			t.Errorf("stdout %q lacks %q", f.stdout.String(), resolved)
		}
	})
	t.Run("relative path", func(t *testing.T) {
		t.Chdir(filepath.Dir(dir))
		f := &fake{stdinTTY: true}
		if f.run("--hyperlinks=always", filepath.Join(filepath.Base(dir), "doc.md")); !strings.Contains(f.stdout.String(), resolved) {
			t.Errorf("stdout %q lacks %q", f.stdout.String(), resolved)
		}
	})
	t.Run("directory argument", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
		f := &fake{stdinTTY: true}
		if f.run("--hyperlinks=always", dir); !strings.Contains(f.stdout.String(), resolved) {
			t.Errorf("stdout %q lacks %q", f.stdout.String(), resolved)
		}
	})
	t.Run("piped stdin", func(t *testing.T) {
		f := &fake{stdin: doc}
		if f.run("--hyperlinks=always"); !strings.Contains(f.stdout.String(), asWritten) {
			t.Errorf("stdout %q lacks %q", f.stdout.String(), asWritten)
		}
	})
	t.Run("explicit stdin", func(t *testing.T) {
		f := &fake{stdin: doc, stdinTTY: true}
		if f.run("--hyperlinks=always", "-"); !strings.Contains(f.stdout.String(), asWritten) {
			t.Errorf("stdout %q lacks %q", f.stdout.String(), asWritten)
		}
	})
	t.Run("hyperlinks disabled", func(t *testing.T) {
		f := &fake{stdinTTY: true}
		if f.run("--hyperlinks=never", path); !strings.Contains(f.stdout.String(), "guide (docs/guide.md)") {
			t.Errorf("stdout %q", f.stdout.String())
		}
	})
}

func TestVersion(t *testing.T) {
	path := writeFile(t, "# From file\n")
	for _, args := range [][]string{{"--version"}, {"--version", path}} {
		f := &fake{}
		e := f.env()
		e.version = "v1.2.3"
		stdin := strings.NewReader("# From stdin\n")
		e.stdin = stdin
		fileRead := false
		e.readFile = func(string) ([]byte, error) {
			fileRead = true
			return nil, errors.New("unexpected read")
		}
		if code := run(args, e); code != 0 {
			t.Errorf("%v: exit %d, want 0", args, code)
		}
		if got := f.stdout.String(); got != "termd v1.2.3\n" {
			t.Errorf("%v: stdout %q, want %q", args, got, "termd v1.2.3\n")
		}
		if fileRead || stdin.Len() != len("# From stdin\n") {
			t.Errorf("%v: input was read", args)
		}
	}
}

func TestPipeOutput(t *testing.T) {
	f := &fake{stdin: "**bold** [link](https://example.com)\n\n" + longParagraph + "\n```go\nfunc main() {}\n```\n"}
	if code := f.run(); code != 0 {
		t.Fatalf("exit %d", code)
	}
	out := f.stdout.String()
	if strings.ContainsRune(out, 0x1b) {
		t.Errorf("piped output contains ESC: %q", out)
	}
	if w := maxLineWidth(out); w > 80 || w < 70 {
		t.Errorf("piped output is %d wide, want wrapping at 80", w)
	}
	if !strings.Contains(out, "link (https://example.com)") {
		t.Errorf("link not in plain form: %q", out)
	}
	if f.detectCalls != 0 {
		t.Errorf("terminal queried %d times in a pipe", f.detectCalls)
	}
}

func TestPipeOutputWithControlCharacters(t *testing.T) {
	f := &fake{stdin: "hello \x1b[31mRED\x1b[0m and \x1b]0;pwned\x07 title\n"}
	if code := f.run(); code != 0 {
		t.Fatalf("exit %d", code)
	}
	out := f.stdout.String()
	if strings.ContainsAny(out, "\x1b\x07") {
		t.Errorf("piped output contains ESC or BEL: %q", out)
	}
	for _, want := range []string{"␛[31m", "␛]0;pwned␇"} {
		if !strings.Contains(out, want) {
			t.Errorf("piped output %q lacks %q", out, want)
		}
	}
}

func TestTerminalOutput(t *testing.T) {
	t.Run("terminal width and styles", func(t *testing.T) {
		f := &fake{stdin: "**bold**\n\n" + longParagraph, stdoutTTY: true, width: 100, height: 1000}
		f.run()
		out := f.stdout.String()
		if !strings.Contains(out, "\x1b[1mbold") {
			t.Errorf("no styles in terminal output: %q", out)
		}
		if w := maxLineWidth(stripSGR(out)); w > 100 || w < 90 {
			t.Errorf("output is %d wide, want wrapping at 100", w)
		}
	})
	t.Run("width flag overrides the terminal", func(t *testing.T) {
		f := &fake{stdin: longParagraph, stdoutTTY: true, width: 100, height: 1000}
		f.run("--width", "60")
		if w := maxLineWidth(f.stdout.String()); w > 60 || w < 50 {
			t.Errorf("output is %d wide, want wrapping at 60", w)
		}
	})
	t.Run("hyperlinks forced in a pipe", func(t *testing.T) {
		f := &fake{stdin: "[docs](https://example.com)\n"}
		f.run("--hyperlinks=always")
		out := f.stdout.String()
		if !strings.Contains(out, "\x1b]8;;https://example.com\x1b\\docs") || strings.Contains(out, "\x1b[") {
			t.Errorf("forced hyperlinks: %q", out)
		}
	})
	t.Run("hyperlinks disabled in a terminal", func(t *testing.T) {
		f := &fake{stdin: "[docs](https://example.com)\n", stdoutTTY: true, width: 80, height: 100}
		f.run("--hyperlinks=never")
		out := f.stdout.String()
		if strings.Contains(out, "\x1b]8;;") || !strings.Contains(stripSGR(out), "docs (https://example.com)") {
			t.Errorf("hyperlinks never: %q", out)
		}
	})
}

func stripSGR(s string) string {
	var b strings.Builder
	for len(s) > 0 {
		if strings.HasPrefix(s, "\x1b[") {
			end := strings.IndexByte(s, 'm')
			if end < 0 {
				break
			}
			s = s[end+1:]
			continue
		}
		b.WriteByte(s[0])
		s = s[1:]
	}
	return b.String()
}

const goBlock = "```go\nfunc main() {}\n```\n"

func TestThemeSelection(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		doc       string
		stdoutTTY bool
		light     bool
		calls     int
	}{
		{"auto queries once", nil, goBlock + "\n" + goBlock, true, false, 1},
		{"explicit dark does not query", []string{"--theme=dark"}, goBlock, true, false, 0},
		{"explicit light does not query", []string{"--theme=light"}, goBlock, true, false, 0},
		{"no code does not query", nil, "# Title\n", true, false, 0},
		{"pipe does not query", nil, goBlock, false, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fake{stdin: tt.doc, stdoutTTY: tt.stdoutTTY, width: 80, height: 100, lightOK: true, light: tt.light}
			f.run(tt.args...)
			if f.detectCalls != tt.calls {
				t.Errorf("terminal queried %d times, want %d", f.detectCalls, tt.calls)
			}
		})
	}
}

func TestDetectedThemeIsUsed(t *testing.T) {
	render := func(args []string, light, ok bool) string {
		f := &fake{stdin: goBlock, stdoutTTY: true, width: 80, height: 100, colorterm: "truecolor", light: light, lightOK: ok}
		f.run(args...)
		return f.stdout.String()
	}
	lightOut := render([]string{"--theme=light"}, false, false)
	darkOut := render([]string{"--theme=dark"}, false, false)
	if lightOut == darkOut {
		t.Fatal("light and dark themes produce the same output")
	}
	if got := render(nil, true, true); got != lightOut {
		t.Error("light background did not select the light theme")
	}
	if got := render(nil, false, true); got != darkOut {
		t.Error("dark background did not select the dark theme")
	}
	if got := render(nil, false, false); got != darkOut {
		t.Error("no answer did not fall back to the dark theme")
	}
}

func TestColorDepth(t *testing.T) {
	for _, tt := range []struct {
		colorterm string
		want      string
		not       string
	}{
		{"truecolor", "38;2;", "38;5;"},
		{"24bit", "38;2;", "38;5;"},
		{"", "38;5;", "38;2;"},
	} {
		f := &fake{stdin: goBlock, stdoutTTY: true, width: 80, height: 100, colorterm: tt.colorterm}
		f.run("--theme=dark")
		out := f.stdout.String()
		if !strings.Contains(out, tt.want) || strings.Contains(out, tt.not) {
			t.Errorf("COLORTERM=%q: output %q", tt.colorterm, out)
		}
	}
}

func TestPagerPath(t *testing.T) {
	found := func(string) (string, error) { return "/usr/bin/less", nil }
	missing := func(string) (string, error) { return "", errors.New("not found") }
	long := strings.Repeat("line\n", 50)
	wide := "short\n" + strings.Repeat("x", 100) + "\n"
	styledWide := "\x1b[1m" + strings.Repeat("x", 70) + "\x1b[0m \x1b]8;;https://example.com\x1b\\docs\x1b]8;;\x1b\\\n"
	tabbed := "\t\t\t\t\t\t\t\t\tx\n"
	tests := []struct {
		name    string
		tty     bool
		sizeOK  bool
		width   int
		height  int
		out     string
		noPager bool
		look    func(string) (string, error)
		want    bool
	}{
		{"long output in a terminal", true, true, 80, 24, long, false, found, true},
		{"fits the terminal", true, true, 80, 60, long, false, found, false},
		{"short but wider than the terminal", true, true, 80, 24, wide, false, found, true},
		{"escapes do not count as width", true, true, 80, 24, styledWide, false, found, false},
		{"tabs advance to multiples of 8", true, true, 70, 24, tabbed, false, found, true},
		{"tabs fit a wider terminal", true, true, 80, 24, tabbed, false, found, false},
		{"not a terminal", false, true, 80, 24, long, false, found, false},
		{"no-pager flag", true, true, 80, 24, long, true, found, false},
		{"less missing", true, true, 80, 24, long, false, missing, false},
		{"unknown size", true, false, 0, 0, long, false, found, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, ok := pagerPath(tt.tty, tt.sizeOK, tt.width, tt.height, tt.out, tt.noPager, tt.look)
			if ok != tt.want || (ok && path != "/usr/bin/less") {
				t.Errorf("pagerPath = (%q, %v), want %v", path, ok, tt.want)
			}
		})
	}
}

func TestPaging(t *testing.T) {
	long := strings.Repeat("paragraph\n\n", 40)

	f := &fake{stdin: long, stdoutTTY: true, width: 80, height: 24, hasLess: true}
	if code := f.run(); code != 0 || len(f.pagerCalls) != 1 || f.stdout.Len() != 0 {
		t.Errorf("long document: exit %d, pager calls %d, stdout %d bytes", code, len(f.pagerCalls), f.stdout.Len())
	}

	f = &fake{stdin: "short\n", stdoutTTY: true, width: 80, height: 24, hasLess: true}
	if f.run(); len(f.pagerCalls) != 0 || f.stdout.Len() == 0 {
		t.Errorf("short document was paged")
	}

	f = &fake{stdin: long, stdoutTTY: true, width: 80, height: 24, hasLess: true}
	if f.run("--no-pager"); len(f.pagerCalls) != 0 || f.stdout.Len() == 0 {
		t.Errorf("--no-pager document was paged")
	}

	f = &fake{stdin: long, stdoutTTY: true, width: 80, height: 24}
	if code := f.run(); code != 0 || f.stdout.Len() == 0 {
		t.Errorf("without less: exit %d, stdout %d bytes", code, f.stdout.Len())
	}

	wideTable := "| a | b |\n|---|---|\n| " + strings.Repeat("x", 60) + " | " + strings.Repeat("y", 60) + " |\n"
	f = &fake{stdin: wideTable, stdoutTTY: true, width: 80, height: 24, hasLess: true}
	if f.run(); len(f.pagerCalls) != 1 || f.stdout.Len() != 0 {
		t.Errorf("short document with a wide table was not paged")
	}
}

func TestUnsupportedDiagramExitsZero(t *testing.T) {
	path := writeFile(t, "```mermaid\nstateDiagram-v2\n    [*] --> Idle\n```\n\nAfter.\n")
	f := &fake{stdinTTY: true}
	if code := f.run(path); code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	out := f.stdout.String()
	if !strings.Contains(out, "stateDiagram-v2 - not supported") || !strings.Contains(out, "After.") {
		t.Errorf("output %q", out)
	}
}

func TestInvalidFrontmatterExitsZero(t *testing.T) {
	path := writeFile(t, "---\ntitle: [unclosed\n---\n\nAfter.\n")
	f := &fake{stdinTTY: true}
	if code := f.run(path); code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	out := f.stdout.String()
	if !strings.Contains(out, "frontmatter - invalid YAML") || !strings.Contains(out, "After.") {
		t.Errorf("output %q", out)
	}
}

const (
	goFile   = "package main\n\n// main does nothing.\nfunc main() {}\n"
	yamlFile = "---\n# Server settings\nport: 8080\n"
	twoLines = "first line\nsecond line\n"
)

func TestCodeFiles(t *testing.T) {
	tests := []struct {
		name, file, content, want string
	}{
		{"recognized extension", "main.go", goFile, goFile},
		{"recognized whole name", "Makefile", "build:\n\tgo build ./...\n", "build:\n\tgo build ./...\n"},
		{"comment lines are not headings", "config.yaml", yamlFile, yamlFile},
		{"no line break at the end", "main.go", "func main() {}", "func main() {}\n"},
		{"empty file", "empty.go", "", ""},
		{"markdown file", "doc.md", twoLines, "first line second line\n"},
		{"plain text file", "notes.txt", twoLines, "first line second line\n"},
		{"unrecognized name", "README", twoLines, "first line second line\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fake{stdinTTY: true}
			if code := f.run(writeNamedFile(t, tt.file, tt.content)); code != 0 || f.stdout.String() != tt.want {
				t.Errorf("exit %d, stdout %q, want %q", code, f.stdout.String(), tt.want)
			}
		})
	}
}

func TestStdinIsMarkdown(t *testing.T) {
	for _, args := range [][]string{nil, {"-"}} {
		f := &fake{stdin: "# Server settings\nport: 8080\n"}
		want := "# Server settings\n\nport: 8080\n"
		if code := f.run(args...); code != 0 || f.stdout.String() != want {
			t.Errorf("%v: exit %d, stdout %q, want %q", args, code, f.stdout.String(), want)
		}
	}
}

func TestCodeFileInTerminal(t *testing.T) {
	f := &fake{stdoutTTY: true, width: 80, height: 24, colorterm: "truecolor", lightOK: true}
	f.run(writeNamedFile(t, "main.go", goFile))
	if out := f.stdout.String(); !strings.Contains(out, "\x1b[38;2;") || stripSGR(out) != goFile {
		t.Errorf("main.go in a terminal: %q", out)
	}
	if f.detectCalls != 1 {
		t.Errorf("terminal queried %d times, want 1", f.detectCalls)
	}

	wide := strings.Repeat("x", 120)
	f = &fake{stdoutTTY: true, width: 80, height: 24, hasLess: true}
	f.run(writeNamedFile(t, "main.go", "package main\n\n// "+wide+"\n"))
	if len(f.pagerCalls) != 1 || !strings.Contains(stripSGR(f.pagerCalls[0]), "// "+wide+"\n") {
		t.Errorf("a code file with a wide line was not paged in full: %q", f.pagerCalls)
	}
}
