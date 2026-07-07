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
func lineCount(b []byte) int {
	n := 0
	if l := len(b); l > 0 {
		n = 1
		for i := 0; i < l; i++ {
			if b[i] == lf {
				n++
			}
		}
		if b[l-1] == lf {
			return n - 1
		}
	}
	return n
}

// nextLine returns the first line and remaining data, trimming LF and preceding CR characters.
// Line terminators are excluded from the returned line.
func nextLine(b []byte) (line []byte, rest []byte) {
	if l := len(b); l > 0 {
		for i := 0; i < l; i++ {
			if b[i] == lf {
				end := i
				for i > 0 && isSpace(b[i-1]) {
					i--
				}
				if i > 0 {
					if end+1 < l {
						return b[:i], b[end+1:]
					}
					return b[:i], nil
				}
				if end+1 < l {
					return nil, b[end+1:]
				}
				return nil, nil
			}
		}
		return b, nil
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

// stripPrefixIfSuffixes strips leading prefix bytes and returns the matching
// suffix if the remaining data exactly matches one of the allowed suffixes.
func stripPrefixIfSuffixes(b []byte, prefix byte, suffixes [][]byte) []byte {
	bufLn := len(b)
	for _, suffix := range suffixes {
		if bytes.HasSuffix(b, suffix) {
			sfxLn := len(suffix)
			pos := bufLn - sfxLn
			for pos > 0 && b[pos-1] == prefix {
				pos--
			}
			if pos == 0 {
				return b[bufLn-sfxLn:]
			}
			return nil
		}
	}
	return nil
}

// stripPrefix removes all leading prefix bytes and returns a subslice of the original data
func stripPrefix(b []byte, prefix byte) []byte {
	i, l := 0, len(b)
	for i < l && b[i] == prefix {
		i++
	}
	return b[i:]
}

// keywordValue extracts a supported keyword and its optional value,
// allowing the keyword to be prefixed with comment characters ('#')
func keywordValue(b []byte) (keyword, value []byte) {
	i, l := 0, len(b)

	// Skip leading space
	for i < l && isSpace(b[i]) {
		i++
	}

	// Keyword that comes first
	start := i
	for i < l && !isSpace(b[i]) {
		i++
	}
	if start == i {
		return nil, nil
	}
	keyword = stripPrefixIfSuffixes(b[start:i], cmntEol, [][]byte{
		[]byte("device"), []byte("makeoptions"), []byte("options"),
	})
	if keyword == nil {
		return nil, nil
	}

	// Skip space
	for i < l && isSpace(b[i]) {
		i++
	}

	// Value that comes second
	start = i
	for i < l && !isSpace(b[i]) {
		i++
	}
	if start != i {
		value = b[start:i]
	}

	return keyword, value
}

// concat efficiently joins multiple byte slices into a single string
func concat(parts ...[]byte) string {
	l, n := len(parts), 0
	for i := 0; i < l; i++ {
		n += len(parts[i])
	}

	b := make([]byte, 0, n)
	for i := 0; i < l; i++ {
		b = append(b, parts[i]...)
	}

	return string(b)
}
