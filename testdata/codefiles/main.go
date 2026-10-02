// Command greet prints a greeting for every name given on the command line.
package main

import (
	"fmt"
	"os"
	"strings"
)

/*
The usage text is markdown:

```
# greet
greet NAME...
```
*/
const usage = "Usage: greet NAME..."

// greeting returns the greeting for name.
func greeting(name string) string {
	return fmt.Sprintf("hello, %s", strings.TrimSpace(name))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	for _, name := range os.Args[1:] {
		fmt.Println(greeting(name)) // this comment makes the line wider than eighty columns
	}
}
