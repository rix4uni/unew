package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "v1.0.0"

func main() {
	var appendMode bool
	var quietMode bool
	var trim bool
	var showVersion bool

	flag.BoolVar(&appendMode, "a", false, "append output; do not sort")
	flag.BoolVar(&quietMode, "q", false, "quiet mode (no output at all on terminal)")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace before comparison")
	flag.BoolVar(&showVersion, "v", false, "print version information and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("unew version:", version)
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
		var f *os.File
		var err error
		if fn != "" {
			f, err = os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				f = nil
			}
		}
		defer func() {
			if f != nil {
				f.Close()
			}
		}()

		w := bufio.NewWriter(f)
		defer w.Flush()

		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			if trim {
				line = strings.TrimSpace(line)
			}
			if !quietMode {
				fmt.Println(line)
			}
			if f != nil {
				w.WriteString(line + "\n")
			}
		}
		return
	}

	if fn != "" {
		var f *os.File
		var err error
		if fn != "" {
			f, err = os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				f = nil
			}
		}
		defer func() {
			if f != nil {
				f.Close()
			}
		}()

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
				if f != nil {
					w.WriteString(line + "\n")
				}
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
				if !quietMode {
					fmt.Println(line)
				}
			}
		}
	}
}
