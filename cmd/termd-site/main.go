// Command termd-site builds the termd landing page into a directory. With
// -shots, it writes the pages for the theme screenshots of the README
// instead.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ekalinin/termd/internal/site"
)

func main() {
	shots := flag.Bool("shots", false, "write one page per theme for the README screenshots instead of the landing page")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: termd-site [-shots] DIR")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	build := site.Build
	if *shots {
		build = site.BuildShots
	}
	if err := build(flag.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "termd-site: %v\n", err)
		os.Exit(1)
	}
}
