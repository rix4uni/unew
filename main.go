package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var quietMode bool
	var dryRun bool
	var trim bool
	flag.BoolVar(&quietMode, "q", false, "quiet mode (no output at all)")
	flag.BoolVar(&dryRun, "d", false, "don't append anything to the file, just print the new lines to stdout")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace before comparison")
	flag.Parse()

	fn := flag.Arg(0)

	lines := make(map[string]struct{}) // Use struct{} for values to save space

	if fn != "" {
		// read the whole file into a map if it exists
		r, err := os.Open(fn)
		if err == nil {
			sc := bufio.NewScanner(r)

			for sc.Scan() {
				line := sc.Text()
				if trim {
					line = strings.TrimSpace(line)
				}
				lines[line] = struct{}{}
			}
			r.Close()
		}

		if !dryRun {
			// re-open the file for appending new stuff
			var err error
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to open file for writing: %s\n", err)
				return
			}
			defer f.Close()

			// Create a buffered writer to minimize the number of write system calls
			w := bufio.NewWriter(f)
			defer w.Flush()

			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				line := sc.Text()
				if trim {
					line = strings.TrimSpace(line)
				}

				if _, exists := lines[line]; !exists {
					lines[line] = struct{}{}
					if !quietMode {
						fmt.Println(line)
					}
					if !dryRun && fn != "" {
						w.WriteString(line + "\n")
					}
				}
			}
		}
	}
}
