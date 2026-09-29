// Command termd-site builds the termd landing page into a directory.
package main

import (
	"fmt"
	"os"

	"github.com/ekalinin/termd/internal/site"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: termd-site DIR")
		os.Exit(2)
	}
	if err := site.Build(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "termd-site: %v\n", err)
		os.Exit(1)
	}
}
