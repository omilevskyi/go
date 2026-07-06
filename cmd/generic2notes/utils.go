package main

import (
	"bytes"
)

const (
	ht = '\t' // HT (Horizontal Tab)
	lf = '\n' // LF (Line Feed)
	vt = '\v' // VT (Vertical Tab)
	ff = '\f' // FF (Form Feed)
	cr = '\r' // CR (Carriage Return)
	sp = ' '  // Space

	cmntEol = '#'
)

// lineCount returns the number of lines in data without allocating,
// allowing accurate preallocation of a slice for subsequent line parsing.
func lineCount(data []byte) int {
	n := 0
	if l := len(data); l > 0 {
		n = 1
		for i := 0; i < l; i++ {
			if data[i] == lf {
				n++
			}
		}
		if data[l-1] == lf {
			return n - 1
		}
	}
	return n
}

// nextLine returns the first line and remaining data, trimming LF and preceding CR characters.
// Line terminators are excluded from the returned line.
func nextLine(data []byte) (line []byte, rest []byte) {
	if l := len(data); l > 0 {
		for i := 0; i < l; i++ {
			if data[i] == lf {
				end := i
				for i > 0 && data[i-1] == cr {
					i--
				}
				if i > 0 {
					if end+1 < l {
						return data[:i], data[end+1:]
					}
					return data[:i], nil
				}
				if end+1 < l {
					return nil, data[end+1:]
				}
				return nil, nil
			}
		}
		return data, nil
	}
	return nil, nil
}

// isSpace reports whether c is an ASCII whitespace character
func isSpace(c byte) bool {
	switch c {
	case sp, ht, lf, cr, ff, vt:
		return true
	}
	return false
}

// func uniqueBytes() []byte {
// 	return strconv.AppendInt(nil, time.Now().UnixNano(), 36) // 0-9 + a-z
// }

// stripPrefixIfSuffixes strips leading prefix bytes and returns the matching
// suffix if the remaining data exactly matches one of the allowed suffixes.
func stripPrefixIfSuffixes(data []byte, prefix byte, suffixes [][]byte) []byte {
	ldata := len(data)
	for _, suffix := range suffixes {
		if bytes.HasSuffix(data, suffix) {
			lsfx := len(suffix)
			pos := ldata - lsfx
			for pos > 0 && data[pos-1] == prefix {
				pos--
			}
			if pos == 0 {
				return data[ldata-lsfx:]
			}
			return nil
		}
	}
	return nil
}

// keywordValue extracts a supported keyword and its optional value,
// allowing the keyword to be prefixed with comment characters ('#')
func keywordValue(data []byte) (keyword, value []byte) {
	i, l := 0, len(data)

	// Skip leading space
	for i < l && isSpace(data[i]) {
		i++
	}

	// Keyword that comes first
	start := i
	for i < l && !isSpace(data[i]) {
		i++
	}
	if start == i {
		return nil, nil
	}
	keyword = stripPrefixIfSuffixes(data[start:i], cmntEol, [][]byte{
		[]byte("device"), []byte("makeoptions"), []byte("options"),
	})
	if keyword == nil {
		return nil, nil
	}

	// Skip space
	for i < l && isSpace(data[i]) {
		i++
	}

	// Value that comes second
	start = i
	for i < l && !isSpace(data[i]) {
		i++
	}
	if start != i {
		value = data[start:i]
	}

	return keyword, value
}
