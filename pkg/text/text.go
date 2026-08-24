package text

import "bytes"

// White space constants
const (
	SP = ' '  // 0x20 Space
	HT = '\t' // 0x09 HT (Horizontal Tab)
	LF = '\n' // 0x0a LF (Line Feed)
	CR = '\r' // 0x0d CR (Carriage Return)
	FF = '\f' // 0x0c FF (Form Feed)
	VT = '\v' // 0x0b VT (Vertical Tab)
)

// Text -
type Text struct {
	SkipEmptyLine  bool
	SkipWhiteSpace bool
	TrimLeadSpace  bool
	TrimTrailSpace bool
	WhiteSpaces    []byte
}

// Option -
type Option func(*Text)

// WithSkipEmptyLine -
func WithSkipEmptyLine(v bool) Option {
	return func(t *Text) {
		t.SkipEmptyLine = v
	}
}

// WithSkipWhiteSpace -
func WithSkipWhiteSpace(v bool) Option {
	return func(t *Text) {
		t.SkipWhiteSpace = v
	}
}

// WithWhiteSpaces -
func WithWhiteSpaces(ws []byte) Option {
	return func(t *Text) {
		t.WhiteSpaces = ws
	}
}

// WithTrimLeadSpace -
func WithTrimLeadSpace(v bool) Option {
	return func(t *Text) {
		t.TrimLeadSpace = v
	}
}

// WithTrimTrailSpace -
func WithTrimTrailSpace(v bool) Option {
	return func(t *Text) {
		t.TrimTrailSpace = v
	}
}

// New -
func New(opts ...Option) *Text {
	t := &Text{
		WhiteSpaces: []byte{SP, HT, LF, CR, FF, VT},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// IsSpace reports whether b is a whitespace character
func (t *Text) IsSpace(b byte) bool {
	for _, w := range t.WhiteSpaces {
		if b == w {
			return true
		}
	}
	return false
}

// IsSpaceLine returns true if the line contains only spaces
func (t *Text) IsSpaceLine(line []byte) bool {
	for _, b := range line {
		if !t.IsSpace(b) {
			return false
		}
	}
	return true
}

// LineCount returns the number of lines in data without allocating,
// allowing accurate preallocation of a slice for subsequent line parsing.
func (t *Text) LineCount(b []byte) (count int) {
	if n := len(b); n > 0 {
		var line []byte
		start, length := 0, 0
		for i := 0; i <= n; i++ {
			if i == n || b[i] == LF {
				line, length = b[start:i], i-start
				if (!t.SkipWhiteSpace || (length > 0 && !t.IsSpaceLine(line))) &&
					(!t.SkipEmptyLine || length > 0) {
					count++
				}
				start = i + 1
			}
		}
		if count > 0 && b[n-1] == LF {
			return count - 1
		}
	}
	return count
}

// NextLine returns the first line and remaining data, trimming LF and preceding CR characters.
// Line terminators are excluded from the returned line.
func (t *Text) NextLine(b []byte) (line, rest []byte) {
	for len(b) > 0 {
		if eol := bytes.IndexByte(b, LF); eol >= 0 {
			line = b[:eol]
			if eol+1 < len(b) {
				rest = b[eol+1:]
			} else {
				rest = nil
			}
		} else {
			line, rest = b, nil
		}
		if t.TrimTrailSpace {
			for len(line) > 0 && t.IsSpace(line[len(line)-1]) {
				line = line[:len(line)-1]
			}
		}
		if t.TrimLeadSpace {
			for len(line) > 0 && t.IsSpace(line[0]) {
				line = line[1:]
			}
		}
		if (!t.SkipWhiteSpace || !t.IsSpaceLine(line)) &&
			(!t.SkipEmptyLine || len(line) > 0) { // !(t.SkipEmptyLine && len(line) == 0 || t.SkipWhiteSpace && t.IsSpaceLine(line))
			return line, rest
		}
		b = rest
	}
	return nil, nil
}

// NextLineNoSlice -
func (t *Text) NextLineNoSlice(b []byte) ([]byte, []byte) {
	var start, end, eol int
	for len(b) > 0 {
		start, end, eol = 0, len(b), bytes.IndexByte(b, LF)
		if eol >= 0 {
			end = eol
		}
		if t.TrimTrailSpace {
			for start < end && t.IsSpace(b[end-1]) {
				end--
			}
		}
		if t.TrimLeadSpace {
			for start < end && t.IsSpace(b[start]) {
				start++
			}
		}
		if (!t.SkipWhiteSpace || !t.IsSpaceLine(b[start:end])) &&
			(!t.SkipEmptyLine || start < end) {
			if eol >= 0 && eol+1 < len(b) {
				return b[start:end], b[eol+1:]
			}
			return b[start:end], nil
		}
		if eol < 0 {
			break
		}
		b = b[eol+1:]
	}
	return nil, nil
}
