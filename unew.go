package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const version = "v0.0.4"

func main() {
	// Ignore SIGPIPE to avoid broken pipe errors
	signal.Ignore(syscall.SIGPIPE)

	var appendMode bool
	var quietMode bool
	var trim bool
	var showVersion bool
	var ignoreCase bool
	var stopEmptyFiles bool
	var shuffle bool
	var removeEmptyLines bool
	var split int

	flag.BoolVar(&appendMode, "a", false, "append output; do not sort")
	flag.BoolVar(&quietMode, "q", false, "quiet mode (no output at all on terminal)")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace before comparison")
	flag.BoolVar(&ignoreCase, "i", false, "ignore case during comparison")
	flag.BoolVar(&stopEmptyFiles, "ef", false, "do not create empty files")
	flag.BoolVar(&showVersion, "v", false, "print version information and exit")
	flag.BoolVar(&shuffle, "shuf", false, "shuffle the output lines randomly")
	flag.BoolVar(&removeEmptyLines, "el", false, "remove empty lines from input")
	flag.IntVar(&split, "split", 0, "split the output into files with a specified number of lines per file")
	flag.Parse()

	// Validate flags: if -a is used with any flag other than -q, print an error and exit
	if appendMode && (trim || ignoreCase || stopEmptyFiles || shuffle || showVersion || removeEmptyLines || split > 0) {
		fmt.Println("-q flag is the only flag allowed with -a flag")
		return
	}

	if showVersion {
		fmt.Println("unew version:", version)
		return
	}

	fn := flag.Arg(0)
	if stopEmptyFiles && fn == "" {
		fmt.Println("A filename must be provided with -ef flag")
		return
	}

	if quietMode && fn == "" {
		fmt.Println("A filename must be provided with -q flag")
		return
	}

	lines := make(map[string]struct{}) // Use struct{} for values to save space

	if fn != "" && !appendMode {
		// Read the whole file into a map if it exists
		r, err := os.Open(fn)
		if err == nil {
			sc := bufio.NewScanner(r)
			for sc.Scan() {
				line := sc.Text()
				if trim {
					line = strings.TrimSpace(line)
				}
				if ignoreCase {
					line = strings.ToLower(line)
				}
				lines[line] = struct{}{}
			}
			r.Close()
		}
	}

	// Initialize variables to check if any lines are written
	anyLinesWritten := false
	lineSlice := []string{} // For shuffling purposes

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
			if removeEmptyLines && line == "" {
				continue
			}
			if !quietMode {
				fmt.Println(line)
			}
			if f != nil {
				w.WriteString(line + "\n")
				anyLinesWritten = true
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			os.Remove(fn)
		}
		return
	}

	// Handling the case when splitting lines into different files
	if split > 0 {
		sc := bufio.NewScanner(os.Stdin)
		var fileIndex int
		var lineCount int
		var currentFile *os.File
		var err error
		defer func() {
			if currentFile != nil {
				currentFile.Close()
			}
		}()

		openNewFile := func() {
			if currentFile != nil {
				currentFile.Close()
			}
			fileIndex++
			baseName := strings.TrimSuffix(fn, ".txt") // Add this line to remove .txt if present
			
			// Add .txt suffix only if it was originally present
    		if strings.HasSuffix(fn, ".txt") {
				currentFile, err = os.Create(fmt.Sprintf("%s%d.txt", baseName, fileIndex))
			} else {
				currentFile, err = os.Create(fmt.Sprintf("%s%d", baseName, fileIndex))
			}
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
		}

		openNewFile()
		w := bufio.NewWriter(currentFile)
		defer w.Flush()

		for sc.Scan() {
			line := sc.Text()
			if removeEmptyLines && line == "" {
				continue
			}
			if trim {
				line = strings.TrimSpace(line)
			}
			if ignoreCase {
				line = strings.ToLower(line)
			}
			if _, exists := lines[line]; !exists {
				lines[line] = struct{}{}
				lineSlice = append(lineSlice, line) // Collect lines for shuffling
				if !quietMode && !shuffle {
					fmt.Println(line)
				}
				if currentFile != nil {
					w.WriteString(line + "\n")
					anyLinesWritten = true
					lineCount++
				}

				// Open a new file if the current one reaches the split limit
				if lineCount >= split {
					w.Flush()
					lineCount = 0
					openNewFile()
					w = bufio.NewWriter(currentFile)
				}
			}
		}

		// Flush the writer one last time to ensure all remaining lines are written
		if currentFile != nil {
			w.Flush()
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(lineSlice), func(i, j int) {
				lineSlice[i], lineSlice[j] = lineSlice[j], lineSlice[i]
			})

			for _, line := range lineSlice {
				if !quietMode {
					fmt.Println(line)
				}
				if currentFile != nil {
					w.WriteString(line + "\n")
					anyLinesWritten = true
				}
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			os.Remove(fmt.Sprintf("%s%d.txt", fn, fileIndex))
		}
		return
	}

	// Standard non-split operation
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
			if removeEmptyLines && line == "" {
				continue
			}
			if trim {
				line = strings.TrimSpace(line)
			}
			if ignoreCase {
				line = strings.ToLower(line)
			}
			if _, exists := lines[line]; !exists {
				lines[line] = struct{}{}
				lineSlice = append(lineSlice, line) // Collect lines for shuffling
				if !quietMode && !shuffle {
					fmt.Println(line)
				}
				if f != nil && !shuffle {
					w.WriteString(line + "\n")
					anyLinesWritten = true
				}
			}
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(lineSlice), func(i, j int) {
				lineSlice[i], lineSlice[j] = lineSlice[j], lineSlice[i]
			})

			for _, line := range lineSlice {
				if !quietMode {
					fmt.Println(line)
				}
				if f != nil {
					w.WriteString(line + "\n")
					anyLinesWritten = true
				}
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			os.Remove(fn)
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			if removeEmptyLines && line == "" {
				continue
			}
			if trim {
				line = strings.TrimSpace(line)
			}
			if ignoreCase {
				line = strings.ToLower(line)
			}
			if _, exists := lines[line]; !exists {
				lines[line] = struct{}{}
				lineSlice = append(lineSlice, line) // Collect lines for shuffling
				if !quietMode && !shuffle {
					fmt.Println(line)
				}
			}
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(lineSlice), func(i, j int) {
				lineSlice[i], lineSlice[j] = lineSlice[j], lineSlice[i]
			})

			for _, line := range lineSlice {
				if !quietMode {
					fmt.Println(line)
				}
			}
		}
	}
}
