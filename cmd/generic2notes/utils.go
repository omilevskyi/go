package main

import (
	"bytes"

	txt "github.com/omilevskyi/go/pkg/text"
)

const cmntEol = '#'

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
	i, l, t := 0, len(b), txt.New()

	// Skip leading space
	for i < l && t.IsSpace(b[i]) {
		i++
	}

	// Keyword that comes first
	start := i
	for i < l && !t.IsSpace(b[i]) {
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
	for i < l && t.IsSpace(b[i]) {
		i++
	}

	// Value that comes second
	start = i
	for i < l && !t.IsSpace(b[i]) {
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
	for i := range l {
		n += len(parts[i])
	}

	b := make([]byte, 0, n)
	for i := range l {
		b = append(b, parts[i]...)
	}

	return string(b)
}
