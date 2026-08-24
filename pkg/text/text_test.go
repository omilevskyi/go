package text

import (
	"bytes"
	"reflect"
	"testing"
)

func TestText_LineCount(t *testing.T) {
	tests := []struct {
		name string
		text string
		opts []Option
		want int
	}{
		{
			name: "nil/empty buffer",
			text: "",
			want: 0,
		},
		{
			name: "single line",
			text: "abc",
			want: 1,
		},
		{
			name: "single line with trailing lf",
			text: "abc\n",
			want: 1,
		},
		{
			name: "multiple lines",
			text: "a\nb\nc",
			want: 3,
		},
		{
			name: "multiple lines trailing lf",
			text: "a\nb\nc\n",
			want: 3,
		},
		{
			name: "only empty lines",
			text: "\n\n\n",
			want: 3,
		},
		{
			name: "skip empty lines",
			text: "\n\n\n",
			opts: []Option{
				WithSkipEmptyLine(true),
			},
			want: 0,
		},
		{
			name: "empty line in middle",
			text: "a\n\nb",
			want: 3,
		},
		{
			name: "skip empty line in middle",
			text: "a\n\nb",
			opts: []Option{
				WithSkipEmptyLine(true),
			},
			want: 2,
		},
		{
			name: "whitespace line only spaces",
			text: "a\n   \nb",
			want: 3,
		},
		{
			name: "skip whitespace line spaces",
			text: "a\n   \nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 2,
		},
		{
			name: "skip whitespace line tabs",
			text: "a\n\t\t\nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 2,
		},
		{
			name: "skip whitespace mixed",
			text: "a\n \t \r \nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 2,
		},
		{
			name: "skip whitespace treats empty as skipped",
			text: "a\n\nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 2,
		},
		{
			name: "both options enabled",
			text: "a\n\n\t\n \n\nb",
			opts: []Option{
				WithSkipEmptyLine(true),
				WithSkipWhiteSpace(true),
			},
			want: 2,
		},
		{
			name: "all whitespace lines",
			text: " \n\t\n\r\n",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 0,
		},
		{
			name: "custom whitespace set",
			text: "a\nxxx\nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
				WithWhiteSpaces([]byte{'x'}),
			},
			want: 2,
		},
		{
			name: "custom whitespace set partial",
			text: "a\nxxy\nb",
			opts: []Option{
				WithSkipWhiteSpace(true),
				WithWhiteSpaces([]byte{'x'}),
			},
			want: 3,
		},
		{
			name: "single empty line",
			text: "\n",
			want: 1,
		},
		{
			name: "single empty line skipped",
			text: "\n",
			opts: []Option{
				WithSkipEmptyLine(true),
			},
			want: 0,
		},
		{
			name: "single whitespace line skipped",
			text: "   ",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 0,
		},
		{
			name: "single non whitespace line",
			text: " abc ",
			opts: []Option{
				WithSkipWhiteSpace(true),
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.opts...).LineCount([]byte(tt.text))
			if got != tt.want {
				t.Fatalf(
					"LineCount(%q) = %d, want %d",
					tt.text,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestText_IsSpace(t *testing.T) {
	txt := New()

	tests := []struct {
		b    byte
		want bool
	}{
		{' ', true},
		{'\t', true},
		{'\n', true},
		{'a', false},
		{'0', false},
	}

	for _, tt := range tests {
		if got := txt.IsSpace(tt.b); got != tt.want {
			t.Fatalf("IsSpace(%q)=%v want %v", tt.b, got, tt.want)
		}
	}
}

func TestText_IsSpaceLine(t *testing.T) {
	txt := New()

	tests := []struct {
		line string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"\t\t", true},
		{" \t\r ", true},
		{"a", false},
		{" a ", false},
	}

	for _, tt := range tests {
		if got := txt.IsSpaceLine([]byte(tt.line)); got != tt.want {
			t.Fatalf("IsSpaceLine(%q)=%v want %v", tt.line, got, tt.want)
		}
	}
}

func collectLines(tk *Text, b []byte) [][]byte {
	var lines [][]byte

	for {
		line, rest := tk.NextLine(b)
		if line == nil && rest == nil {
			break
		}

		lines = append(lines, append([]byte(nil), line...))
		b = rest
	}

	return lines
}

func TestNextLine(t *testing.T) {
	tests := []struct {
		name string
		text string
		cfg  *Text
		want []string
	}{
		{
			name: "empty input",
			text: "",
			cfg:  New(),
			want: nil,
		},
		{
			name: "single line",
			text: "abc",
			cfg:  New(),
			want: []string{"abc"},
		},
		{
			name: "single line trailing lf",
			text: "abc\n",
			cfg:  New(),
			want: []string{"abc"},
		},
		{
			name: "multiple lines",
			text: "a\nb\nc",
			cfg:  New(),
			want: []string{"a", "b", "c"},
		},
		{
			name: "multiple lines trailing lf",
			text: "a\nb\nc\n",
			cfg:  New(),
			want: []string{"a", "b", "c"},
		},
		{
			name: "preserve empty lines",
			text: "a\n\nb",
			cfg:  New(),
			want: []string{"a", "", "b"},
		},
		{
			name: "preserve leading empty lines",
			text: "\n\nabc",
			cfg:  New(),
			want: []string{"", "", "abc"},
		},
		{
			name: "preserve trailing empty lines",
			text: "abc\n\n",
			cfg:  New(),
			want: []string{"abc", ""},
		},
		{
			name: "skip empty lines",
			text: "a\n\nb\n\n\nc",
			cfg: New(
				WithSkipEmptyLine(true),
			),
			want: []string{"a", "b", "c"},
		},
		{
			name: "only empty lines",
			text: "\n\n\n",
			cfg:  New(),
			want: []string{"", "", ""},
		},
		{
			name: "only empty lines skipped",
			text: "\n\n\n",
			cfg: New(
				WithSkipEmptyLine(true),
			),
			want: nil,
		},
		{
			name: "whitespace lines preserved",
			text: "a\n   \n\t\t\nb",
			cfg:  New(),
			want: []string{"a", "   ", "\t\t", "b"},
		},
		{
			name: "skip whitespace lines",
			text: "a\n   \n\t\t\nb",
			cfg: New(
				WithSkipWhiteSpace(true),
			),
			want: []string{"a", "b"},
		},
		{
			name: "skip whitespace and empty",
			text: "a\n\n \n\t\nb",
			cfg: New(
				WithSkipEmptyLine(true),
				WithSkipWhiteSpace(true),
			),
			want: []string{"a", "b"},
		},
		{
			name: "trim leading spaces",
			text: "  a\n\tb\n c",
			cfg: New(
				WithTrimLeadSpace(true),
			),
			want: []string{"a", "b", "c"},
		},
		{
			name: "trim trailing spaces",
			text: "a  \nb\t\nc ",
			cfg: New(
				WithTrimTrailSpace(true),
			),
			want: []string{"a", "b", "c"},
		},
		{
			name: "trim both sides",
			text: "  a  \n\tb\t\n c ",
			cfg: New(
				WithTrimLeadSpace(true),
				WithTrimTrailSpace(true),
			),
			want: []string{"a", "b", "c"},
		},
		{
			name: "trim both keeps inner spaces",
			text: "  hello world  ",
			cfg: New(
				WithTrimLeadSpace(true),
				WithTrimTrailSpace(true),
			),
			want: []string{"hello world"},
		},
		{
			name: "trim converts whitespace line to empty",
			text: "   \n\t\t\nabc",
			cfg: New(
				WithTrimLeadSpace(true),
				WithTrimTrailSpace(true),
			),
			want: []string{"", "", "abc"},
		},
		{
			name: "trim and skip empty",
			text: "   \n\t\t\nabc",
			cfg: New(
				WithTrimLeadSpace(true),
				WithTrimTrailSpace(true),
				WithSkipEmptyLine(true),
			),
			want: []string{"abc"},
		},
		{
			name: "trim and skip whitespace",
			text: "   \n\t\t\nabc",
			cfg: New(
				WithTrimLeadSpace(true),
				WithTrimTrailSpace(true),
				WithSkipWhiteSpace(true),
			),
			want: []string{"abc"},
		},
		{
			name: "custom whitespace x",
			text: "xxx\nabc\nxxxx",
			cfg: New(
				WithSkipWhiteSpace(true),
				WithWhiteSpaces([]byte{'x'}),
			),
			want: []string{"abc"},
		},
		{
			name: "custom whitespace partial",
			text: "xxx\nxax\nxxxx",
			cfg: New(
				WithSkipWhiteSpace(true),
				WithWhiteSpaces([]byte{'x'}),
			),
			want: []string{"xax"},
		},
		{
			name: "last line without lf",
			text: "a\nb\nc",
			cfg:  New(),
			want: []string{"a", "b", "c"},
		},
		{
			name: "single whitespace line",
			text: " ",
			cfg:  New(),
			want: []string{" "},
		},
		{
			name: "single whitespace line skipped",
			text: " ",
			cfg: New(
				WithSkipWhiteSpace(true),
			),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLines := collectLines(tt.cfg, []byte(tt.text))

			var got []string
			for _, line := range gotLines {
				got = append(got, string(line))
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf(
					"lines=%#v want=%#v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestNextLine_Rest(t *testing.T) {
	txt := New()

	line, rest := txt.NextLine([]byte("abc\ndef\nghi"))

	if !bytes.Equal(line, []byte("abc")) {
		t.Fatalf("line=%q want=%q", line, "abc")
	}

	if !bytes.Equal(rest, []byte("def\nghi")) {
		t.Fatalf("rest=%q want=%q", rest, "def\nghi")
	}
}

func TestNextLine_SkipMany(t *testing.T) {
	txt := New(
		WithSkipEmptyLine(true),
	)

	line, rest := txt.NextLine([]byte("\n\n\nabc"))

	if string(line) != "abc" {
		t.Fatalf("got %q want %q", line, "abc")
	}

	if rest != nil {
		t.Fatalf("rest must be nil")
	}
}

func TestLines(t *testing.T) {
	txt := New()

	var got [][]byte

	for line := range txt.Lines([]byte("abc\ndef\nghi")) {
		got = append(got, line)
	}

	want := [][]byte{
		[]byte("abc"),
		[]byte("def"),
		[]byte("ghi"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%q want=%q", got, want)
	}
}
