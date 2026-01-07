package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const version = "v0.0.8"

// parseSize parses a size string (e.g., "1GB", "500MB", "1024KB", "512B") and converts it to bytes
func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("size string cannot be empty")
	}

	// Convert to uppercase for case-insensitive matching
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Find the unit (B, KB, MB, GB)
	var unit string
	var numberStr string

	if strings.HasSuffix(sizeStr, "GB") {
		unit = "GB"
		numberStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		unit = "MB"
		numberStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "KB") {
		unit = "KB"
		numberStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "B") {
		unit = "B"
		numberStr = strings.TrimSuffix(sizeStr, "B")
	} else {
		return 0, fmt.Errorf("invalid size format: missing unit (B, KB, MB, GB)")
	}

	// Parse the number
	numberStr = strings.TrimSpace(numberStr)
	if numberStr == "" {
		return 0, fmt.Errorf("invalid size format: missing number")
	}

	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %v", err)
	}

	if number <= 0 {
		return 0, fmt.Errorf("size must be greater than 0")
	}

	// Convert to bytes
	var bytes int64
	switch unit {
	case "B":
		bytes = int64(number)
	case "KB":
		bytes = int64(number * 1024)
	case "MB":
		bytes = int64(number * 1024 * 1024)
	case "GB":
		bytes = int64(number * 1024 * 1024 * 1024)
	default:
		return 0, fmt.Errorf("unsupported unit: %s (supported: B, KB, MB, GB)", unit)
	}

	return bytes, nil
}

// processLine processes a line according to the given flags
// Returns the processed line and whether it should be skipped (true = skip)
func processLine(line string, trim, ignoreCase, removeEmptyLines bool) (string, bool) {
	if removeEmptyLines && line == "" {
		return "", true
	}
	if trim {
		line = strings.TrimSpace(line)
	}
	if ignoreCase {
		line = strings.ToLower(line)
	}
	return line, false
}

// generateFileName generates a filename with the given base name, file index, and suffix handling
func generateFileName(baseName string, fileIndex int, hasTxtSuffix bool) string {
	if hasTxtSuffix {
		return fmt.Sprintf("%s%d.txt", baseName, fileIndex)
	}
	return fmt.Sprintf("%s%d", baseName, fileIndex)
}

// shuffleLines shuffles the given lines and writes them to the writer/file
// Returns whether any lines were written and any error
func shuffleLines(lines []string, quietMode bool, writer *bufio.Writer, file *os.File) (bool, error) {
	if len(lines) == 0 {
		return false, nil
	}

	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	anyLinesWritten := false
	for _, line := range lines {
		if !quietMode {
			fmt.Println(line)
		}
		if writer != nil && file != nil {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return anyLinesWritten, err
			}
			anyLinesWritten = true
		}
	}
	return anyLinesWritten, nil
}

// ensureParentDir ensures the parent directory of the given path exists.
func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

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
	var divide int
	var sizeLimit string

	flag.BoolVar(&appendMode, "a", false, "append lines without deduplication (can be combined with other flags)")
	flag.BoolVar(&quietMode, "q", false, "suppress all output to terminal/stdout")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace from each line before processing")
	flag.BoolVar(&ignoreCase, "i", false, "treat uppercase and lowercase as identical when comparing lines")
	flag.BoolVar(&stopEmptyFiles, "ef", false, "do not create output files if no lines are written")
	flag.BoolVar(&showVersion, "version", false, "display version number and exit")
	flag.BoolVar(&shuffle, "shuf", false, "randomly shuffle the output lines before writing")
	flag.BoolVar(&removeEmptyLines, "el", false, "skip empty lines from input")
	flag.IntVar(&split, "split", 0, "split output into multiple files, each containing specified number of lines (requires filename prefix)")
	flag.IntVar(&divide, "divide", 0, "divide input into N equal files, distributing lines evenly (requires filename prefix, N must be >= 2)")
	flag.StringVar(&sizeLimit, "size", "", "split output into multiple files based on size limit (requires filename prefix, e.g., 1GB, 500MB, 1024KB, 512B)")
	flag.Parse()

	// Validate -divide flag: must be >= 2
	if divide > 0 && divide < 2 {
		fmt.Println("-divide flag must be >= 2")
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

	if divide > 0 && fn == "" {
		fmt.Println("A filename prefix must be provided with -divide flag")
		return
	}

	if sizeLimit != "" && fn == "" {
		fmt.Println("A filename prefix must be provided with -size flag")
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
				processedLine, skip := processLine(line, trim, ignoreCase, false) // Don't remove empty lines when reading from file
				if !skip {
					lines[processedLine] = struct{}{}
				}
			}
			r.Close()
		}
	}

	// Initialize variables to check if any lines are written
	anyLinesWritten := false
	lineSlice := []string{} // For shuffling purposes

	if appendMode && split == 0 && divide == 0 && sizeLimit == "" {
		var f *os.File
		var err error
		if fn != "" {
			if err := ensureParentDir(fn); err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
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
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}
			if shuffle {
				lineSlice = append(lineSlice, processedLine)
			}
			if !shuffle && !quietMode {
				fmt.Println(processedLine)
			}
			if f != nil && !shuffle {
				w.WriteString(processedLine + "\n")
				anyLinesWritten = true
			}
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			shuffledWritten, err := shuffleLines(lineSlice, quietMode, w, f)
			if err == nil && shuffledWritten {
				anyLinesWritten = true
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			fileInfo, err := os.Stat(fn)
			// Check if the file existed and had size > 0 before execution
			fileExisted := err == nil && fileInfo.Size() > 0

			if !anyLinesWritten && !fileExisted {
				os.Remove(fn) // Only remove if it didn't exist or was empty before
			}
		}
		return
	}

	// Handling the case when splitting lines into different files
	if split > 0 {
		baseName := strings.TrimSuffix(fn, ".txt")
		hasTxtSuffix := strings.HasSuffix(fn, ".txt")

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
			fileName := generateFileName(baseName, fileIndex, hasTxtSuffix)
			if err := ensureParentDir(fileName); err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
			currentFile, err = os.Create(fileName)
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
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}

			if appendMode {
				if shuffle {
					lineSlice = append(lineSlice, processedLine) // keep duplicates for shuffle
				}
				if !quietMode && !shuffle {
					fmt.Println(processedLine)
				}
				if currentFile != nil {
					w.WriteString(processedLine + "\n")
					anyLinesWritten = true
					lineCount++
				}
			} else {
				if _, exists := lines[processedLine]; !exists {
					lines[processedLine] = struct{}{}
					if shuffle {
						lineSlice = append(lineSlice, processedLine) // Collect lines for shuffling
					}
					if !quietMode && !shuffle {
						fmt.Println(processedLine)
					}
					if currentFile != nil {
						w.WriteString(processedLine + "\n")
						anyLinesWritten = true
						lineCount++
					}
				}
			}

			// Open a new file if the current one reaches the split limit
			if lineCount >= split {
				w.Flush()
				lineCount = 0
				openNewFile()
				w = bufio.NewWriter(currentFile)
			}
		}

		// Flush the writer one last time to ensure all remaining lines are written
		if currentFile != nil {
			w.Flush()
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			shuffledWritten, _ := shuffleLines(lineSlice, quietMode, w, currentFile)
			if shuffledWritten {
				anyLinesWritten = true
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			fileName := generateFileName(baseName, fileIndex, hasTxtSuffix)
			os.Remove(fileName)
		}
		return
	}

	// Handling the case when dividing lines into equal files
	if divide > 0 {
		// Streaming mode for append (LOW MEMORY)
		if appendMode {
			baseName := strings.TrimSuffix(fn, ".txt")
			hasTxtSuffix := strings.HasSuffix(fn, ".txt")

			// Open all files upfront
			files := make([]*os.File, divide)
			writers := make([]*bufio.Writer, divide)

			for i := 0; i < divide; i++ {
				fileName := generateFileName(baseName, i+1, hasTxtSuffix)
				if err := ensureParentDir(fileName); err != nil {
					fmt.Printf("Error creating directory for %s: %v\n", fileName, err)
					return
				}

				file, err := os.Create(fileName)
				if err != nil {
					fmt.Printf("Error creating file %s: %v\n", fileName, err)
					return
				}
				files[i] = file
				writers[i] = bufio.NewWriter(file)
			}

			// Ensure all files are closed and flushed
			defer func() {
				for i := 0; i < divide; i++ {
					if writers[i] != nil {
						writers[i].Flush()
					}
					if files[i] != nil {
						files[i].Close()
					}
				}
			}()

			// Stream lines in round-robin fashion
			lineIndex := 0
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				line := sc.Text()
				processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
				if skip {
					continue
				}

				// Write to file in round-robin fashion
				fileIdx := lineIndex % divide
				writers[fileIdx].WriteString(processedLine + "\n")
				if !quietMode {
					fmt.Println(processedLine)
				}
				lineIndex++
			}

			return
		}

		// Non-append mode: Read all lines first for deduplication
		sc := bufio.NewScanner(os.Stdin)
		allLines := []string{}
		for sc.Scan() {
			line := sc.Text()
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}

			if _, exists := lines[processedLine]; exists {
				continue
			}
			lines[processedLine] = struct{}{}
			allLines = append(allLines, processedLine)
		}

		totalLines := len(allLines)

		// Validate that divide doesn't exceed line count
		if divide > totalLines {
			fmt.Printf("Error: -divide value (%d) exceeds total line count (%d)\n", divide, totalLines)
			return
		}

		// Calculate lines per file and remainder
		linesPerFile := totalLines / divide
		remainder := totalLines % divide

		// Extract base name from prefix (handle .txt suffix like -split does)
		baseName := strings.TrimSuffix(fn, ".txt")
		hasTxtSuffix := strings.HasSuffix(fn, ".txt")

		// Distribute lines across N files
		lineIndex := 0
		for fileIndex := 1; fileIndex <= divide; fileIndex++ {
			// Determine how many lines this file should get
			linesForThisFile := linesPerFile
			if fileIndex <= remainder {
				linesForThisFile++
			}

			// Create output file
			fileName := generateFileName(baseName, fileIndex, hasTxtSuffix)

			if err := ensureParentDir(fileName); err != nil {
				fmt.Printf("Error creating directory for %s: %v\n", fileName, err)
				return
			}

			file, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", fileName, err)
				return
			}

			w := bufio.NewWriter(file)

			// Write lines to this file
			for i := 0; i < linesForThisFile && lineIndex < totalLines; i++ {
				w.WriteString(allLines[lineIndex] + "\n")
				if !quietMode {
					fmt.Println(allLines[lineIndex])
				}
				lineIndex++
			}

			w.Flush()
			file.Close()
		}
		return
	}

	// Handling the case when splitting lines based on file size
	if sizeLimit != "" {
		// Parse the size limit
		maxSizeBytes, err := parseSize(sizeLimit)
		if err != nil {
			fmt.Printf("Error parsing size limit: %v\n", err)
			return
		}

		// Extract base name from prefix (handle .txt suffix like -split/-divide does)
		baseName := strings.TrimSuffix(fn, ".txt")
		hasTxtSuffix := strings.HasSuffix(fn, ".txt")

		var fileIndex int
		var currentFile *os.File
		var w *bufio.Writer
		var currentFileSize int64
		var err2 error

		// Helper function to create a new file
		openNewFile := func() {
			if currentFile != nil {
				w.Flush()
				currentFile.Close()
			}
			fileIndex++
			fileName := generateFileName(baseName, fileIndex, hasTxtSuffix)

			if err := ensureParentDir(fileName); err != nil {
				fmt.Printf("Error creating directory for %s: %v\n", fileName, err)
				return
			}

			currentFile, err2 = os.Create(fileName)
			if err2 != nil {
				fmt.Printf("Error creating file %s: %v\n", fileName, err2)
				return
			}
			w = bufio.NewWriter(currentFile)
			currentFileSize = 0
		}

		// Open the first file
		openNewFile()

		defer func() {
			if currentFile != nil {
				w.Flush()
				currentFile.Close()
			}
		}()

		// Read lines from stdin and write to files based on size
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}

			// Calculate line size (including newline character)
			lineSize := int64(len(processedLine) + 1) // +1 for the newline character

			if appendMode {
				// If adding this line would exceed the size limit, create a new file
				// (but only if current file already has content, to allow single large lines)
				if currentFileSize > 0 && currentFileSize+lineSize > maxSizeBytes {
					openNewFile()
				}

				// Write line to current file without duplicate checking
				if currentFile != nil {
					w.WriteString(processedLine + "\n")
					currentFileSize += lineSize
					if !quietMode {
						fmt.Println(processedLine)
					}
				}
			} else {
				// Check for duplicates
				if _, exists := lines[processedLine]; !exists {
					lines[processedLine] = struct{}{}

					// If adding this line would exceed the size limit, create a new file
					// (but only if current file already has content, to allow single large lines)
					if currentFileSize > 0 && currentFileSize+lineSize > maxSizeBytes {
						openNewFile()
					}

					// Write line to current file
					if currentFile != nil {
						w.WriteString(processedLine + "\n")
						currentFileSize += lineSize
						if !quietMode {
							fmt.Println(processedLine)
						}
					}
				}
			}
		}

		// Flush the last file
		if currentFile != nil {
			w.Flush()
		}
		return
	}

	// Standard non-split operation
	if fn != "" {
		var f *os.File
		var err error
		if fn != "" {
			if err := ensureParentDir(fn); err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
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
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}
			if _, exists := lines[processedLine]; !exists {
				lines[processedLine] = struct{}{}
				if shuffle {
					lineSlice = append(lineSlice, processedLine) // Collect lines for shuffling
				}
				if !quietMode && !shuffle {
					fmt.Println(processedLine)
				}
				if f != nil && !shuffle {
					w.WriteString(processedLine + "\n")
					anyLinesWritten = true
				}
			}
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			shuffledWritten, _ := shuffleLines(lineSlice, quietMode, w, f)
			if shuffledWritten {
				anyLinesWritten = true
			}
		}

		// Handle empty file creation based on -ef flag
		if stopEmptyFiles && fn != "" && !anyLinesWritten {
			fileInfo, err := os.Stat(fn)
			// Check if the file existed and had size > 0 before execution
			fileExisted := err == nil && fileInfo.Size() > 0

			if !anyLinesWritten && !fileExisted {
				os.Remove(fn) // Only remove if it didn't exist or was empty before
			}
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			processedLine, skip := processLine(line, trim, ignoreCase, removeEmptyLines)
			if skip {
				continue
			}
			if _, exists := lines[processedLine]; !exists {
				lines[processedLine] = struct{}{}
				if shuffle {
					lineSlice = append(lineSlice, processedLine) // Collect lines for shuffling
				}
				if !quietMode && !shuffle {
					fmt.Println(processedLine)
				}
			}
		}

		// Handle shuffling if the -shuf flag is set
		if shuffle {
			shuffleLines(lineSlice, quietMode, nil, nil)
		}
	}
}
