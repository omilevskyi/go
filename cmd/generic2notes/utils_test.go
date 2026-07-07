package main

import (
	"bytes"
	"testing"
)

func TestLineCount(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "nil",
			data: nil,
			want: 0,
		},
		{
			name: "empty",
			data: []byte{},
			want: 0,
		},
		{
			name: "single empty line",
			data: []byte(""),
			want: 0,
		},
		{
			name: "single line without lf",
			data: []byte("abc"),
			want: 1,
		},
		{
			name: "single char",
			data: []byte("a"),
			want: 1,
		},
		{
			name: "single line with trailing lf",
			data: []byte("abc\n"),
			want: 1,
		},
		{
			name: "single empty line with lf",
			data: []byte("\n"),
			want: 1,
		},
		{
			name: "two lines",
			data: []byte("abc\ndef"),
			want: 2,
		},
		{
			name: "two lines trailing lf",
			data: []byte("abc\ndef\n"),
			want: 2,
		},
		{
			name: "three lines",
			data: []byte("a\nb\nc"),
			want: 3,
		},
		{
			name: "three lines trailing lf",
			data: []byte("a\nb\nc\n"),
			want: 3,
		},
		{
			name: "empty line in middle",
			data: []byte("a\n\nb"),
			want: 3,
		},
		{
			name: "empty first line",
			data: []byte("\na"),
			want: 2,
		},
		{
			name: "empty last line",
			data: []byte("a\n"),
			want: 1,
		},
		{
			name: "multiple empty lines",
			data: []byte("\n\n\n"),
			want: 3,
		},
		{
			name: "only lfs count as lines",
			data: []byte("\n\n"),
			want: 2,
		},
		{
			name: "crlf",
			data: []byte("a\r\nb\r\nc"),
			want: 3,
		},
		{
			name: "crlf trailing",
			data: []byte("a\r\nb\r\nc\r\n"),
			want: 3,
		},
		{
			name: "only crlf",
			data: []byte("\r\n"),
			want: 1,
		},
		{
			name: "multiple crlf",
			data: []byte("\r\n\r\n\r\n"),
			want: 3,
		},
		{
			name: "cr without lf",
			data: []byte("a\rb\rc"),
			want: 1,
		},
		{
			name: "mixed",
			data: []byte("a\n\r\nb\n\nc\r\n"),
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lineCount(tt.data)
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNextLine(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		line []byte
		end  []byte
	}{
		{
			name: "nil",
			data: nil,
			line: nil,
			end:  nil,
		},
		{
			name: "empty",
			data: []byte{},
			line: nil,
			end:  nil,
		},
		{
			name: "single line without newline",
			data: []byte("abc"),
			line: []byte("abc"),
			end:  nil,
		},
		{
			name: "single line with lf",
			data: []byte("abc\n"),
			line: []byte("abc"),
			end:  nil,
		},
		{
			name: "single char with lf",
			data: []byte("a\n"),
			line: []byte("a"),
			end:  nil,
		},
		{
			name: "only lf",
			data: []byte("\n"),
			line: nil,
			end:  nil,
		},
		{
			name: "empty line then data",
			data: []byte("\nabc"),
			line: nil,
			end:  []byte("abc"),
		},
		{
			name: "two lines",
			data: []byte("abc\ndef"),
			line: []byte("abc"),
			end:  []byte("def"),
		},
		{
			name: "two lines with trailing lf",
			data: []byte("abc\ndef\n"),
			line: []byte("abc"),
			end:  []byte("def\n"),
		},
		{
			name: "empty middle line",
			data: []byte("abc\n\ndef"),
			line: []byte("abc"),
			end:  []byte("\ndef"),
		},
		{
			name: "crlf",
			data: []byte("abc\r\ndef"),
			line: []byte("abc"),
			end:  []byte("def"),
		},
		{
			name: "crlf ending file",
			data: []byte("abc\r\n"),
			line: []byte("abc"),
			end:  nil,
		},
		{
			name: "empty crlf line",
			data: []byte("\r\nabc"),
			line: nil,
			end:  []byte("abc"),
		},
		{
			name: "multiple cr before lf",
			data: []byte("abc\r\r\r\ndef"),
			line: []byte("abc"),
			end:  []byte("def"),
		},
		{
			name: "line consists only of cr",
			data: []byte("\r\n"),
			line: nil,
			end:  nil,
		},
		{
			name: "multiple empty lines",
			data: []byte("\n\n\n"),
			line: nil,
			end:  []byte("\n\n"),
		},
		{
			name: "lf at beginning",
			data: []byte("\nabc\ndef"),
			line: nil,
			end:  []byte("abc\ndef"),
		},
		{
			name: "crlf at beginning",
			data: []byte("\r\nabc\ndef"),
			line: nil,
			end:  []byte("abc\ndef"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, end := nextLine(tt.data)

			if !bytes.Equal(line, tt.line) {
				t.Fatalf(
					"line mismatch:\n got=%q\nwant=%q",
					line,
					tt.line,
				)
			}

			if !bytes.Equal(end, tt.end) {
				t.Fatalf(
					"end mismatch:\n got=%q\nwant=%q",
					end,
					tt.end,
				)
			}
		})
	}
}

func TestNextLineIterative(t *testing.T) {
	data := []byte("a\r\n\nb\nc\r\r\r\n")

	var lines [][]byte

	for len(data) > 0 {
		var line []byte
		line, data = nextLine(data)
		lines = append(lines, line)
	}

	want := [][]byte{
		[]byte("a"),
		nil,
		[]byte("b"),
		[]byte("c"),
	}

	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}

	for i := range want {
		if !bytes.Equal(lines[i], want[i]) {
			t.Fatalf(
				"line %d: got=%q want=%q",
				i,
				lines[i],
				want[i],
			)
		}
	}
}

func TestNextLineConsumesExactlyOneLine(t *testing.T) {
	input := []byte("abc\r\nxyz")

	line, end := nextLine(input)

	if !bytes.Equal(line, []byte("abc")) {
		t.Fatalf("unexpected line: %q", line)
	}

	if !bytes.Equal(end, []byte("xyz")) {
		t.Fatalf("unexpected end: %q", end)
	}
}

func TestLineCountMatchesNextLine(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		[]byte("abc"),
		[]byte("abc\n"),
		[]byte("\n"),
		[]byte("\n\n"),
		[]byte("a\nb\nc"),
		[]byte("a\nb\nc\n"),
		[]byte("a\r\nb\r\nc"),
		[]byte("a\r\n\r\nc"),
		[]byte("\r\n\r\n\r\n"),
		[]byte("a\n\r\nb\n\nc\r\n"),
	}

	for _, input := range tests {
		count := lineCount(input)

		var lines int
		data := input

		for len(data) > 0 {
			lines++
			_, data = nextLine(data)
		}

		if count != lines {
			t.Fatalf(
				"input=%q lineCount=%d nextLine=%d",
				input,
				count,
				lines,
			)
		}
	}
}

func TestStripPrefixIfSuffixes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		prefix   byte
		suffixes [][]byte
		want     []byte
	}{
		{
			name:     "nil data",
			data:     nil,
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:     "empty data",
			data:     []byte{},
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:     "no suffixes",
			data:     []byte("###device"),
			prefix:   '#',
			suffixes: nil,
			want:     nil,
		},
		{
			name:     "exact suffix",
			data:     []byte("device"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     []byte("device"),
		},
		{
			name:     "single prefix",
			data:     []byte("#device"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     []byte("device"),
		},
		{
			name:     "multiple prefixes",
			data:     []byte("#####device"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     []byte("device"),
		},
		{
			name:     "different prefix byte",
			data:     []byte("----device"),
			prefix:   '-',
			suffixes: [][]byte{[]byte("device")},
			want:     []byte("device"),
		},
		{
			name:     "text before prefixes",
			data:     []byte("foo###device"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:     "text before suffix",
			data:     []byte("foodevice"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:     "prefix inside text",
			data:     []byte("foo#device"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:     "suffix not at end",
			data:     []byte("###devicexxx"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("device")},
			want:     nil,
		},
		{
			name:   "another suffix",
			data:   []byte("###disk"),
			prefix: '#',
			suffixes: [][]byte{
				[]byte("device"),
				[]byte("disk"),
			},
			want: []byte("disk"),
		},
		{
			name:   "first matching suffix wins",
			data:   []byte("###device"),
			prefix: '#',
			suffixes: [][]byte{
				[]byte("device"),
				[]byte("ice"),
			},
			want: []byte("device"),
		},
		{
			name:   "later suffix matches",
			data:   []byte("###ice"),
			prefix: '#',
			suffixes: [][]byte{
				[]byte("device"),
				[]byte("ice"),
			},
			want: []byte("ice"),
		},
		{
			name:     "empty suffix matches empty string",
			data:     []byte(""),
			prefix:   '#',
			suffixes: [][]byte{[]byte("")},
			want:     []byte(""),
		},
		{
			name:     "prefixes with empty suffix",
			data:     []byte("###"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("")},
			want:     []byte(""),
		},
		{
			name:     "text before empty suffix",
			data:     []byte("foo"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("")},
			want:     nil,
		},
		{
			name:     "single byte suffix",
			data:     []byte("####x"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("x")},
			want:     []byte("x"),
		},
		{
			name:     "suffix longer than data",
			data:     []byte("abc"),
			prefix:   '#',
			suffixes: [][]byte{[]byte("abcdef")},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPrefixIfSuffixes(
				tt.data,
				tt.prefix,
				tt.suffixes,
			)

			if !bytes.Equal(got, tt.want) {
				t.Fatalf(
					"got=%q want=%q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestStripPrefixIfSuffixesReturnsOriginalSlice(t *testing.T) {
	data := []byte("###device")

	got := stripPrefixIfSuffixes(
		data,
		'#',
		[][]byte{[]byte("device")},
	)

	if !bytes.Equal(got, []byte("device")) {
		t.Fatalf("unexpected result: %q", got)
	}

	expected := data[len(data)-len("device"):]

	if len(got) > 0 && &got[0] != &expected[0] {
		t.Fatal("returned slice is not backed by original buffer")
	}
}

func TestKeywordValue(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		keyword []byte
		value   []byte
	}{
		{
			name:    "nil",
			input:   nil,
			keyword: nil,
			value:   nil,
		},
		{
			name:    "empty",
			input:   []byte{},
			keyword: nil,
			value:   nil,
		},
		{
			name:    "spaces only",
			input:   []byte("   \t\r\n"),
			keyword: nil,
			value:   nil,
		},

		// valid keywords
		{
			name:    "device only",
			input:   []byte("device"),
			keyword: []byte("device"),
			value:   nil,
		},
		{
			name:    "makeoptions only",
			input:   []byte("makeoptions"),
			keyword: []byte("makeoptions"),
			value:   nil,
		},
		{
			name:    "options only",
			input:   []byte("options"),
			keyword: []byte("options"),
			value:   nil,
		},

		// keyword + value
		{
			name:    "device value",
			input:   []byte("device foo"),
			keyword: []byte("device"),
			value:   []byte("foo"),
		},
		{
			name:    "makeoptions value",
			input:   []byte("makeoptions BAR"),
			keyword: []byte("makeoptions"),
			value:   []byte("BAR"),
		},
		{
			name:    "options value",
			input:   []byte("options XENHVM"),
			keyword: []byte("options"),
			value:   []byte("XENHVM"),
		},

		// whitespace handling
		{
			name:    "leading spaces",
			input:   []byte("   options value"),
			keyword: []byte("options"),
			value:   []byte("value"),
		},
		{
			name:    "multiple spaces",
			input:   []byte("options     value"),
			keyword: []byte("options"),
			value:   []byte("value"),
		},
		{
			name:    "tabs",
			input:   []byte("\toptions\tvalue"),
			keyword: []byte("options"),
			value:   []byte("value"),
		},
		{
			name:    "mixed whitespace",
			input:   []byte(" \t options \t value "),
			keyword: []byte("options"),
			value:   []byte("value"),
		},

		// commented keywords
		{
			name:    "commented device",
			input:   []byte("#device foo"),
			keyword: []byte("device"),
			value:   []byte("foo"),
		},
		{
			name:    "multiply commented device",
			input:   []byte("#####device foo"),
			keyword: []byte("device"),
			value:   []byte("foo"),
		},
		{
			name:    "commented options",
			input:   []byte("##options XENHVM"),
			keyword: []byte("options"),
			value:   []byte("XENHVM"),
		},
		{
			name:    "commented makeoptions",
			input:   []byte("###makeoptions BAR"),
			keyword: []byte("makeoptions"),
			value:   []byte("BAR"),
		},

		// invalid keywords
		{
			name:    "unknown keyword",
			input:   []byte("foo bar"),
			keyword: nil,
			value:   nil,
		},
		{
			name:    "commented unknown keyword",
			input:   []byte("#foo bar"),
			keyword: nil,
			value:   nil,
		},
		{
			name:    "keyword suffix mismatch",
			input:   []byte("devices foo"),
			keyword: nil,
			value:   nil,
		},
		{
			name:    "keyword prefix mismatch",
			input:   []byte("mydevice foo"),
			keyword: nil,
			value:   nil,
		},
		{
			name:    "text before comment prefix",
			input:   []byte("foo#device bar"),
			keyword: nil,
			value:   nil,
		},

		// value handling
		{
			name:    "empty value",
			input:   []byte("options "),
			keyword: []byte("options"),
			value:   nil,
		},
		{
			name:    "many spaces after keyword",
			input:   []byte("options      "),
			keyword: []byte("options"),
			value:   nil,
		},
		{
			name:    "extra fields ignored",
			input:   []byte("options value another field"),
			keyword: []byte("options"),
			value:   []byte("value"),
		},
		{
			name:    "real example",
			input:   []byte("options    XENHVM          # Xen HVM kernel infrastructure"),
			keyword: []byte("options"),
			value:   []byte("XENHVM"),
		},

		// comment attached to value
		{
			name:    "comment attached to value",
			input:   []byte("options XENHVM#comment"),
			keyword: []byte("options"),
			value:   []byte("XENHVM#comment"),
		},
		{
			name:    "comment separated from value",
			input:   []byte("options XENHVM #comment"),
			keyword: []byte("options"),
			value:   []byte("XENHVM"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyword, value := keywordValue(tt.input)

			if !bytes.Equal(keyword, tt.keyword) {
				t.Fatalf(
					"keyword mismatch:\n got=%q\nwant=%q",
					keyword,
					tt.keyword,
				)
			}

			if !bytes.Equal(value, tt.value) {
				t.Fatalf(
					"value mismatch:\n got=%q\nwant=%q",
					value,
					tt.value,
				)
			}
		})
	}
}

func TestKeywordValueReturnsSlicesOfOriginalBuffer(t *testing.T) {
	data := []byte("options XENHVM")

	keyword, value := keywordValue(data)

	if !bytes.Equal(keyword, []byte("options")) {
		t.Fatalf("unexpected keyword: %q", keyword)
	}

	if !bytes.Equal(value, []byte("XENHVM")) {
		t.Fatalf("unexpected value: %q", value)
	}

	if &keyword[0] != &data[0] {
		t.Fatal("keyword is not backed by original buffer")
	}

	idx := bytes.Index(data, value)
	if &value[0] != &data[idx] {
		t.Fatal("value is not backed by original buffer")
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		prefix byte
		want   []byte
	}{
		{
			name:   "nil",
			input:  nil,
			prefix: '#',
			want:   nil,
		},
		{
			name:   "empty",
			input:  []byte{},
			prefix: '#',
			want:   []byte{},
		},
		{
			name:   "no prefix",
			input:  []byte("device"),
			prefix: '#',
			want:   []byte("device"),
		},
		{
			name:   "single prefix",
			input:  []byte("#device"),
			prefix: '#',
			want:   []byte("device"),
		},
		{
			name:   "multiple prefixes",
			input:  []byte("#####device"),
			prefix: '#',
			want:   []byte("device"),
		},
		{
			name:   "only prefix",
			input:  []byte("#"),
			prefix: '#',
			want:   []byte{},
		},
		{
			name:   "only multiple prefixes",
			input:  []byte("#####"),
			prefix: '#',
			want:   []byte{},
		},
		{
			name:   "different leading character",
			input:  []byte("*device"),
			prefix: '#',
			want:   []byte("*device"),
		},
		{
			name:   "mixed prefixes",
			input:  []byte("##*device"),
			prefix: '#',
			want:   []byte("*device"),
		},
		{
			name:   "zero byte prefix",
			input:  []byte{0, 0, 'a'},
			prefix: 0,
			want:   []byte{'a'},
		},
		{
			name:   "all zero byte prefix",
			input:  []byte{0, 0, 0},
			prefix: 0,
			want:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPrefix(tt.input, tt.prefix)

			if !bytes.Equal(got, tt.want) {
				t.Fatalf(
					"got=%q want=%q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestStripPrefixReturnsOriginalBackingArray(t *testing.T) {
	data := []byte("###device")

	got := stripPrefix(data, '#')

	want := data[3:]

	if !bytes.Equal(got, want) {
		t.Fatalf("got=%q want=%q", got, want)
	}

	if len(got) > 0 && &got[0] != &want[0] {
		t.Fatal("returned slice is not backed by original array")
	}
}

func TestStripPrefixIdempotent(t *testing.T) {
	data := []byte("###device")

	once := stripPrefix(data, '#')
	twice := stripPrefix(once, '#')

	if !bytes.Equal(once, twice) {
		t.Fatalf("once=%q twice=%q", once, twice)
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		name  string
		parts [][]byte
		want  string
	}{
		{
			name:  "no parts",
			parts: nil,
			want:  "",
		},
		{
			name:  "empty slice",
			parts: [][]byte{},
			want:  "",
		},
		{
			name:  "single nil",
			parts: [][]byte{nil},
			want:  "",
		},
		{
			name:  "single empty",
			parts: [][]byte{{}},
			want:  "",
		},
		{
			name:  "single part",
			parts: [][]byte{[]byte("foo")},
			want:  "foo",
		},
		{
			name: "two parts",
			parts: [][]byte{
				[]byte("foo"),
				[]byte("bar"),
			},
			want: "foobar",
		},
		{
			name: "multiple parts",
			parts: [][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
			},
			want: "foobarbaz",
		},
		{
			name: "mixed nil and empty",
			parts: [][]byte{
				nil,
				[]byte("foo"),
				{},
				nil,
				[]byte("bar"),
			},
			want: "foobar",
		},
		{
			name: "contains separator",
			parts: [][]byte{
				[]byte("options"),
				{0},
				[]byte("XENHVM"),
			},
			want: "options\x00XENHVM",
		},
		{
			name: "binary data",
			parts: [][]byte{
				{0x00, 0x01},
				{0xff},
				{0x7f},
			},
			want: string([]byte{
				0x00, 0x01,
				0xff,
				0x7f,
			}),
		},
		{
			name: "utf8",
			parts: [][]byte{
				[]byte("Привет"),
				[]byte(" "),
				[]byte("мир"),
			},
			want: "Привет мир",
		},
		{
			name: "preserve order",
			parts: [][]byte{
				[]byte("3"),
				[]byte("2"),
				[]byte("1"),
			},
			want: "321",
		},
		{
			name: "all empty",
			parts: [][]byte{
				nil,
				{},
				nil,
				{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := concat(tt.parts...)

			if got != tt.want {
				t.Fatalf(
					"got %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestConcatDoesNotModifyInput(t *testing.T) {
	a := []byte("foo")
	b := []byte("bar")

	_ = concat(a, b)

	if !bytes.Equal(a, []byte("foo")) {
		t.Fatalf("a modified: %q", a)
	}

	if !bytes.Equal(b, []byte("bar")) {
		t.Fatalf("b modified: %q", b)
	}
}

func TestConcatManyParts(t *testing.T) {
	parts := make([][]byte, 1000)
	for i := range parts {
		parts[i] = []byte("x")
	}

	got := concat(parts...)

	if len(got) != len(parts) {
		t.Fatalf("got len=%d want=%d", len(got), len(parts))
	}
}
