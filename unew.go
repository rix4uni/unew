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
	var trim bool
	var printMode bool
	var appendMode bool
	flag.BoolVar(&quietMode, "q", false, "quiet mode (no output at all)")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace before comparison")
	flag.BoolVar(&printMode, "p", false, "print only unique input directly to stdout")
	flag.BoolVar(&appendMode, "a", false, "append output in a file and print in teminal")
	flag.Parse()

	fn := flag.Arg(0)

	lines := make(map[string]struct{}) // Use struct{} for values to save space

	if fn != "" && !appendMode {
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
	}

	if appendMode {
		// open the file for appending new stuff
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
			fmt.Println(line) // Printing line to stdout
			w.WriteString(line + "\n") // Appending line to file
		}
		return
	}

	if printMode {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			if trim {
				line = strings.TrimSpace(line)
			}
			if _, exists := lines[line]; !exists {
				lines[line] = struct{}{}
				fmt.Println(line)
			}
		}
		return // exit the program since we've done what was required by the -p flag
	}

	if fn != "" {
		// re-open the file for appending new stuff
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
				w.WriteString(line + "\n")
			}
		}
	}
}
