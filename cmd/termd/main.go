// Command termd renders markdown documents in the terminal.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/rivo/uniseg"
	"golang.org/x/term"

	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/render"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/termbg"
)

const usage = `Usage: termd [flags] [FILE | DIR]

Render a markdown document in the terminal. For a directory DIR, its
README.md, README.markdown or README is rendered. With no FILE, or when FILE
is -, the document is read from standard input.

Flags:
`

// pipeWidth is the output width when stdout is not a terminal.
const pipeWidth = 80

// themeQueryTimeout bounds the wait for the terminal background reply.
const themeQueryTimeout = 100 * time.Millisecond

// env is everything run needs from the outside world, so tests can fake it.
type env struct {
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	stdinTTY  bool
	stdoutTTY bool
	// size returns the terminal size of stdout.
	size     func() (width, height int, ok bool)
	getenv   func(string) string
	readFile func(string) ([]byte, error)
	stat     func(string) (os.FileInfo, error)
	readDir  func(string) ([]os.DirEntry, error)
	lookPath func(string) (string, error)
	// runPager shows content in the pager; it returns an error only when the
	// pager could not be started.
	runPager func(path, content string) error
	// detectLight reports whether the terminal background is light.
	detectLight func() (light, ok bool)
	// version is printed by --version.
	version string
}

func main() {
	os.Exit(run(os.Args[1:], systemEnv()))
}

func systemEnv() env {
	return env{
		stdin:     os.Stdin,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdinTTY:  term.IsTerminal(int(os.Stdin.Fd())),
		stdoutTTY: term.IsTerminal(int(os.Stdout.Fd())),
		size: func() (int, int, bool) {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			return w, h, err == nil && w > 0 && h > 0
		},
		getenv:      os.Getenv,
		readFile:    os.ReadFile,
		stat:        os.Stat,
		readDir:     os.ReadDir,
		lookPath:    exec.LookPath,
		runPager:    runLess,
		detectLight: func() (bool, bool) { return termbg.Light(themeQueryTimeout) },
		version:     buildVersion(),
	}
}

// buildVersion returns the main module version Go recorded in the binary: the
// tag for a release build or go install @tag, a pseudo-version for a build
// from a git checkout, and "(devel)" when no version is recorded.
func buildVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "(devel)"
}

// run executes termd and returns the exit status.
func run(args []string, e env) int {
	fs := flag.NewFlagSet("termd", flag.ContinueOnError)
	fs.SetOutput(e.stderr)
	width := fs.Int("width", 0, "output width in columns (default: terminal width, or 80 when not a terminal)")
	noPager := fs.Bool("no-pager", false, "print directly instead of paging through less")
	hyperlinks := fs.String("hyperlinks", "auto", "terminal hyperlinks: auto, always or never")
	theme := fs.String("theme", "auto", "code highlighting theme: auto, dark or light")
	showVersion := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprint(e.stderr, usage)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *showVersion {
		fmt.Fprintf(e.stdout, "termd %s\n", e.version)
		return 0
	}
	usageError := func(format string, a ...any) int {
		fmt.Fprintf(e.stderr, "termd: "+format+"\n", a...)
		fs.Usage()
		return 2
	}

	widthSet := false
	fs.Visit(func(f *flag.Flag) { widthSet = widthSet || f.Name == "width" })
	if widthSet && *width <= 0 {
		return usageError("--width must be a positive number, got %d", *width)
	}
	if !slices.Contains([]string{"auto", "always", "never"}, *hyperlinks) {
		return usageError("--hyperlinks must be auto, always or never, got %q", *hyperlinks)
	}
	if !slices.Contains([]string{"auto", "dark", "light"}, *theme) {
		return usageError("--theme must be auto, dark or light, got %q", *theme)
	}
	if fs.NArg() > 1 {
		return usageError("expected at most one file, got %d", fs.NArg())
	}

	var file string // the path that is read; empty for stdin
	var src []byte
	var err error
	switch {
	case fs.NArg() == 1 && fs.Arg(0) != "-":
		if file, err = inputPath(fs.Arg(0), e.stat, e.readDir); err == nil {
			src, err = e.readFile(file)
		}
	case fs.NArg() == 1 || !e.stdinTTY:
		src, err = io.ReadAll(e.stdin)
	default:
		fs.Usage()
		return 2
	}
	if err != nil {
		fmt.Fprintf(e.stderr, "termd: %v\n", err)
		return 1
	}

	termWidth, termHeight, sizeOK := 0, 0, false
	if e.stdoutTTY {
		termWidth, termHeight, sizeOK = e.size()
	}
	opts := render.Options{
		Width: outputWidth(*width, widthSet, termWidth, sizeOK),
		Style: style.Options{
			Styled:     e.stdoutTTY,
			Hyperlinks: *hyperlinks == "always" || (*hyperlinks == "auto" && e.stdoutTTY),
			Depth:      colorDepth(e.getenv("COLORTERM")),
		},
		Theme: themeFunc(*theme, e.detectLight),
	}
	out := render.Render(src, opts)

	if path, ok := pagerPath(e.stdoutTTY, sizeOK, termWidth, termHeight, out, *noPager, e.lookPath); ok {
		if err := e.runPager(path, out); err == nil {
			return 0
		}
	}
	io.WriteString(e.stdout, out)
	return 0
}

// readmeNames are the names of a directory's README, compared
// case-insensitively; the first one found is used.
var readmeNames = []string{"README.md", "README.markdown", "README"}

// inputPath returns the file to read for the argument: the argument itself,
// or the README in it when it is a directory. Subdirectories are not searched.
func inputPath(arg string, stat func(string) (os.FileInfo, error), readDir func(string) ([]os.DirEntry, error)) (string, error) {
	if info, err := stat(arg); err != nil || !info.IsDir() {
		// Reading the file reports a missing or unreadable one.
		return arg, nil
	}
	entries, err := readDir(arg)
	if err != nil {
		return "", err
	}
	for _, name := range readmeNames {
		for _, entry := range entries {
			if !strings.EqualFold(entry.Name(), name) {
				continue
			}
			// Only a regular file counts, also through a symbolic link.
			path := filepath.Join(arg, entry.Name())
			if info, err := stat(path); err == nil && info.Mode().IsRegular() {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("no README in %s", arg)
}

// outputWidth picks the layout width: the flag, then the terminal width,
// then the pipe default.
func outputWidth(flagWidth int, flagSet bool, termWidth int, termOK bool) int {
	switch {
	case flagSet:
		return flagWidth
	case termOK:
		return termWidth
	}
	return pipeWidth
}

// colorDepth selects 24-bit colors when COLORTERM advertises them.
func colorDepth(colorterm string) style.Depth {
	switch strings.ToLower(colorterm) {
	case "truecolor", "24bit":
		return style.TrueColor
	}
	return style.Color256
}

// themeFunc returns the lazy theme resolver for the --theme flag. Only
// "auto" queries the terminal, and only when the renderer asks.
func themeFunc(flagTheme string, detectLight func() (bool, bool)) func() highlight.Theme {
	switch flagTheme {
	case "dark":
		return func() highlight.Theme { return highlight.Dark }
	case "light":
		return func() highlight.Theme { return highlight.Light }
	}
	return func() highlight.Theme {
		if light, ok := detectLight(); ok && light {
			return highlight.Light
		}
		return highlight.Dark
	}
}

// pagerPath decides whether to page the output and returns the pager path.
// Output is paged when it is taller or wider than the terminal; without the
// pager the terminal would wrap wide lines and break tables and diagrams.
func pagerPath(stdoutTTY, sizeOK bool, width, height int, out string, noPager bool, lookPath func(string) (string, error)) (string, bool) {
	if !stdoutTTY || !sizeOK || noPager {
		return "", false
	}
	if strings.Count(out, "\n") <= height && widestLine(out) <= width {
		return "", false
	}
	path, err := lookPath("less")
	if err != nil {
		return "", false
	}
	return path, true
}

// escapeRE matches the SGR and OSC 8 sequences termd emits.
var escapeRE = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)

// widestLine returns the display width of the widest line of out, ignoring
// escape sequences. A tab advances to the next multiple of 8 columns, as in
// the terminal.
func widestLine(out string) int {
	widest := 0
	for line := range strings.SplitSeq(escapeRE.ReplaceAllString(out, ""), "\n") {
		col := 0
		g := uniseg.NewGraphemes(line)
		for g.Next() {
			if g.Str() == "\t" {
				col = (col/8 + 1) * 8
				continue
			}
			col += g.Width()
		}
		widest = max(widest, col)
	}
	return widest
}

// runLess shows content in less with raw control sequences and chopped
// long lines, so wide blocks scroll horizontally.
func runLess(path, content string) error {
	cmd := exec.Command(path, "-RS")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if os.Getenv("LESSCHARSET") == "" {
		// The output is always UTF-8, whatever the locale says.
		cmd.Env = append(cmd.Env, "LESSCHARSET=utf-8")
	}
	// Ctrl-C belongs to less while it runs; termd must not exit under it.
	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Wait()
	return nil
}
