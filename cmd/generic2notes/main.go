package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	pathGENERIC  = "/usr/src/sys/amd64/conf/GENERIC"
	pathNOTES    = "/usr/src/sys/conf/NOTES"
	pathNOTES64  = "/usr/src/sys/amd64/conf/NOTES"
	pathDEFAULTS = "/usr/src/sys/amd64/conf/DEFAULTS"

	prefixDEFAULTS string = string(cmntEol) + "DEFAULTS: "

	sep byte = 0
)

func main() {
	genericLines, gnrcMap, err := readConfig(pathGENERIC)
	ut.IsErr(err, 201, "readConfig()")

	_, dfltMap, err := readConfig(pathDEFAULTS)
	ut.IsErr(err, 202, "readConfig()")

	notesLines, err := processNotes(pathNOTES, genericLines, gnrcMap, dfltMap)
	ut.IsErr(err, 203, "processNotes()")

	notes64Lines, err := processNotes(pathNOTES64, genericLines, gnrcMap, dfltMap)
	ut.IsErr(err, 204, "processNotes()")

	err = writeLines(filepath.Base(pathGENERIC), genericLines)
	ut.IsErr(err, 205, "writeLines()")

	err = writeLines(filepath.Base(pathNOTES), notesLines)
	ut.IsErr(err, 206, "writeLines()")

	err = writeLines(filepath.Base(pathNOTES64)+"-AMD64", notes64Lines)
	ut.IsErr(err, 207, "writeLines()")
}

// writeLines writes all lines except those marked as deleted and appends
// an LF terminator after each written line
func writeLines(filepath string, lines [][]byte) error {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return ut.Fringerr(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for i := 0; i < len(lines); i++ {
		// switch {
		// case generics[i] == nil: fmt.Println("<nil>")
		// case len(generics[i]) == 0: fmt.Println("<empty>")
		// default:	fmt.Println(string(lines[i]))
		// }
		if lines[i] == nil || len(lines[i]) > 0 {
			if _, err := w.Write(lines[i]); err != nil {
				return err
			}
			if err := w.WriteByte(lf); err != nil {
				return err
			}
		}
	}

	return nil
}

// readConfig reads GENERIC file, preserves its original lines,
// and builds a lookup map from keyword/value pairs to line indexes
func readConfig(filepath string) ([][]byte, map[string]int, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return nil, nil, ut.Fringerr(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, ut.Fringerr(err)
	}

	lc := lineCount(data)
	lines := make([][]byte, 0, lc)
	m := make(map[string]int)
	for i := 0; i < lc; i++ {
		line, rest := nextLine(data)
		lines = append(lines, line)
		data = rest

		if kwrd, val := keywordValue(line); len(kwrd) > 0 && len(val) > 0 {
			m[concat(kwrd, []byte{sep}, val)] = i
		}
	}
	return lines, m, nil
}

// processNotes reads a notes file, substitutes matching keyword/value pairs
// with their corresponding lines from the generic file, and marks the reused
// generic lines so they can be identified as consumed later
func processNotes(filepath string, generics [][]byte, gnrc, dflt map[string]int) ([][]byte, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return nil, ut.Fringerr(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, ut.Fringerr(err)
	}

	lc := lineCount(data)
	lines := make([][]byte, 0, lc)
	for i := 0; i < lc; i++ {
		line, rest := nextLine(data)

		if kwrd, val := keywordValue(line); len(kwrd) > 0 && len(val) > 0 {
			if ln, ok := gnrc[concat(kwrd, []byte{sep}, val)]; ok {
				line = generics[ln]
				generics[ln] = []byte("") // mark it in such an awkward way that the line has been deleted
			}
			if _, ok := dflt[concat(kwrd, []byte{sep}, val)]; ok {
				line = append([]byte(prefixDEFAULTS), stripPrefix(line, cmntEol)...)
			}
		}

		lines = append(lines, line)
		data = rest
	}
	return lines, nil
}
