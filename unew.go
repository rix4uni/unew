package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var appendMode bool
	var printMode bool
	var quietMode bool
	var trim bool
	flag.BoolVar(&appendMode, "a", false, "append output in a file and print in terminal and (default: filename is required)")
	flag.BoolVar(&printMode, "p", false, "print only unique output and (default: filename is not required)")
	flag.BoolVar(&quietMode, "q", false, "quiet mode (not print output in terminal) and (default: filename is required)")
	flag.BoolVar(&trim, "t", false, "trim whitespace (add unique trim output in a file) and (default: filename is required)")
	flag.Parse()

	// Set trim to true if -q is provided and -t is not set.
	if quietMode && !trim {
		trim = true
	}

	if quietMode && flag.Arg(0) == "" {
		fmt.Fprintf(os.Stderr, "filename is required when using -q flag\n")
		return
	}else if trim && flag.Arg(0) == "" {
		fmt.Fprintf(os.Stderr, "filename is required when using -t flag\n")
		return
	}

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
		if fn == "" {
			fmt.Fprintf(os.Stderr, "filename is required when using -a flag\n")
			return
		}

		f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file for writing: %s\n", err)
			return
		}
		defer f.Close()

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

	if fn != "" {
		// re-open the file for appending new stuff
		f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file for writing: %s\n", err)
			return
		}
		defer f.Close()

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
				if printMode && !quietMode {
					fmt.Println(line)
				}
				w.WriteString(line + "\n")
			}
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			if trim {
				line = strings.TrimSpace(line)
			}
			if _, exists := lines[line]; !exists {
				lines[line] = struct{}{}
				if printMode && !quietMode {
					fmt.Println(line)
				}
			}
		}
	}
}
