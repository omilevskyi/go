package utils

import (
	"reflect"
	"slices"
	"testing"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]int
		expect []string
	}{
		{
			name:   "Empty map",
			input:  map[string]int{},
			expect: nil,
		},
		{
			name: "Single entry",
			input: map[string]int{
				"key1": 1,
			},
			expect: []string{"key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Keys(tt.input)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("Keys() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestKeysMultipleEntries(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]struct{}
	}{
		{
			name: "Multiple entries",
			input: map[string]struct{}{
				"key1": {},
				"key2": {},
				"key3": {},
				"key4": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Keys(tt.input)
			m := make(map[string]struct{}, len(got))
			for _, v := range got {
				m[v] = struct{}{}
			}
			if !reflect.DeepEqual(m, tt.input) {
				t.Errorf("Keys() = %v, want %v", m, tt.input)
			}
		})
	}
}

func TestDistinct(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		got := Distinct([]int{})
		want := []int(nil)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]int{}) = %v; want %v", got, want)
		}
	})

	t.Run("No duplicates", func(t *testing.T) {
		got := Distinct([]string{"a", "b", "c"})
		want := []string{"a", "b", "c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]string{\"a\", \"b\", \"c\"}) = %v; want %v", got, want)
		}
	})

	t.Run("All duplicates", func(t *testing.T) {
		got := Distinct([]int{1, 1, 1, 1})
		want := []int{1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]int{1, 1, 1, 1}) = %v; want %v", got, want)
		}
	})

	t.Run("Mixed duplicates", func(t *testing.T) {
		got := Distinct([]int{1, 2, 1, 3, 2, 4})
		want := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]int{1, 2, 1, 3, 2, 4}) = %v; want %v", got, want)
		}
	})

	t.Run("Single element", func(t *testing.T) {
		got := Distinct([]int{42})
		want := []int{42}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]int{42}) = %v; want %v", got, want)
		}
	})

	t.Run("Strings with duplicates", func(t *testing.T) {
		got := Distinct([]string{"go", "go", "lang", "go", "lang"})
		want := []string{"go", "lang"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]string{\"go\", \"go\", \"lang\", \"go\", \"lang\"}) = %v; want %v", got, want)
		}
	})

	t.Run("Booleans", func(t *testing.T) {
		got := Distinct([]bool{true, false, true, false})
		want := []bool{true, false}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Distinct([]bool{true, false, true, false}) = %v; want %v", got, want)
		}
	})
}

func TestDistinctCopy(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		got := DistinctCopy([]int{})
		want := []int(nil)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]int{}) = %v; want %v", got, want)
		}
	})

	t.Run("No duplicates", func(t *testing.T) {
		got := DistinctCopy([]string{"a", "b", "c"})
		want := []string{"a", "b", "c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]string{\"a\", \"b\", \"c\"}) = %v; want %v", got, want)
		}
	})

	t.Run("All duplicates", func(t *testing.T) {
		got := DistinctCopy([]int{1, 1, 1, 1})
		want := []int{1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]int{1, 1, 1, 1}) = %v; want %v", got, want)
		}
	})

	t.Run("Mixed duplicates", func(t *testing.T) {
		got := DistinctCopy([]int{1, 2, 1, 3, 2, 4})
		want := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]int{1, 2, 1, 3, 2, 4}) = %v; want %v", got, want)
		}
	})

	t.Run("Single element", func(t *testing.T) {
		got := DistinctCopy([]int{42})
		want := []int{42}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]int{42}) = %v; want %v", got, want)
		}
	})

	t.Run("Strings with duplicates", func(t *testing.T) {
		got := DistinctCopy([]string{"go", "go", "lang", "go", "lang"})
		want := []string{"go", "lang"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]string{\"go\", \"go\", \"lang\", \"go\", \"lang\"}) = %v; want %v", got, want)
		}
	})

	t.Run("Booleans", func(t *testing.T) {
		got := DistinctCopy([]bool{true, false, true, false})
		want := []bool{true, false}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DistinctCopy([]bool{true, false, true, false}) = %v; want %v", got, want)
		}
	})
}

func TestDeleteItems(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		targets  []string
		expected []string
		skip     bool
	}{
		{
			name:     "Remove single item",
			input:    []string{"a", "b", "c"},
			targets:  []string{"b"},
			expected: []string{"a", "c"},
			skip:     false,
		},
		{
			name:     "Remove multiple items",
			input:    []string{"a", "b", "c", "b", "d"},
			targets:  []string{"b", "d"},
			expected: []string{"a", "c"},
			skip:     false,
		},
		{
			name:     "Remove all items",
			input:    []string{"x", "x", "x"},
			targets:  []string{"x"},
			expected: []string{},
			skip:     false,
		},
		{
			name:     "Remove none (no match)",
			input:    []string{"a", "b", "c"},
			targets:  []string{"x"},
			expected: []string{"a", "b", "c"},
			skip:     false,
		},
		{
			name:     "Empty input slice",
			input:    []string{},
			targets:  []string{"a"},
			expected: []string{},
			skip:     false,
		},
		{
			name:     "Empty targets",
			input:    []string{"a", "b", "c"},
			targets:  []string{},
			expected: []string{"a", "b", "c"},
			skip:     false,
		},
		{
			name:     "Nil slice pointer",
			input:    nil,
			targets:  []string{"a"},
			expected: nil,
			skip:     true,
		},
		{
			name:     "Remove empty string",
			input:    []string{"", "a", "", "b"},
			targets:  []string{""},
			expected: []string{"a", "b"},
			skip:     false,
		},
		{
			name:     "Duplicates in targets",
			input:    []string{"a", "b", "c", "b"},
			targets:  []string{"b", "b"},
			expected: []string{"a", "c"},
			skip:     false,
		},
		{
			name:     "All elements are targets",
			input:    []string{"a", "b", "c"},
			targets:  []string{"a", "b", "c"},
			expected: []string{},
			skip:     false,
		},
	}

	for _, tt := range tests {
		t.Run("deleteItems()/"+tt.name, func(t *testing.T) {
			if tt.skip { // just check that it doesn't panic
				var nilPtr *[]string = nil
				deleteItems(nilPtr, tt.targets...)
				return
			}

			var inputCopy []string
			if tt.input != nil {
				inputCopy = make([]string, len(tt.input))
				copy(inputCopy, tt.input)
			}
			deleteItems(&inputCopy, tt.targets...)
			if !reflect.DeepEqual(inputCopy, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, inputCopy)
			}
		})
	}
}

func TestDeleteItemsCopy(t *testing.T) {
	type testCase[T comparable] struct {
		name     string
		input    []T
		targets  []T
		expected []T
	}

	intTests := []testCase[int]{
		{"Empty slice", []int{}, []int{1}, nil},
		{"No targets", []int{1, 2, 3}, nil, []int{1, 2, 3}},
		{"Remove one", []int{1, 2, 3}, []int{2}, []int{1, 3}},
		{"Remove multiple", []int{1, 2, 3, 4}, []int{2, 4}, []int{1, 3}},
		{"Remove all", []int{1, 1, 1}, []int{1}, []int{}},
		{"Duplicates in input", []int{1, 2, 2, 3}, []int{2}, []int{1, 3}},
		{"Duplicates in targets", []int{1, 2, 3}, []int{2, 2}, []int{1, 3}},
	}

	stringTests := []testCase[string]{
		{"Remove string", []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{"Remove multiple strings", []string{"a", "b", "c", "d"}, []string{"a", "d"}, []string{"b", "c"}},
		{"Remove nothing", []string{"a", "b"}, []string{"x"}, []string{"a", "b"}},
	}

	type MyInt int
	customTests := []testCase[MyInt]{
		{"Custom type", []MyInt{1, 2, 3}, []MyInt{2}, []MyInt{1, 3}},
	}

	for _, tc := range intTests {
		t.Run("int/"+tc.name, func(t *testing.T) {
			result := deleteItemsCopy(tc.input, tc.targets...)
			if !slices.Equal(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}

	for _, tc := range stringTests {
		t.Run("string/"+tc.name, func(t *testing.T) {
			result := deleteItemsCopy(tc.input, tc.targets...)
			if !slices.Equal(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}

	for _, tc := range customTests {
		t.Run("custom/"+tc.name, func(t *testing.T) {
			result := deleteItemsCopy(tc.input, tc.targets...)
			if !slices.Equal(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Remove empty strings",
			input:    []string{"a", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "All empty strings",
			input:    []string{"", "", ""},
			expected: nil,
		},
		{
			name:     "No empty strings",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Mixed content",
			input:    []string{"a", "", "b", "c", "", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "Empty input slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "Nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Repeated empty strings",
			input:    []string{"a", "", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputCopy []string
			if tt.input != nil {
				inputCopy = make([]string, len(tt.input))
				copy(inputCopy, tt.input)
			}
			result := Compact(inputCopy)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompactCopy(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Remove empty strings",
			input:    []string{"a", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "All empty strings",
			input:    []string{"", "", ""},
			expected: nil,
		},
		{
			name:     "No empty strings",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Mixed content",
			input:    []string{"a", "", "b", "c", "", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "Empty input slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "Nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Repeated empty strings",
			input:    []string{"a", "", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputCopy []string
			if tt.input != nil {
				inputCopy = make([]string, len(tt.input))
				copy(inputCopy, tt.input)
			}
			result := CompactCopy(inputCopy)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTrimQQ(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Common cases
		{`"hello"`, "hello"},
		{`"123"`, "123"},
		{`"a"`, "a"}, // Length is 3, gets removed

		// Without quotes
		{"hello", "hello"},
		{"", ""},
		{"a", "a"},

		// Only one quote

		{`"`, `"`},
		{`"abc`, `"abc`},
		{`abc"`, `abc"`},

		// Asymmetrical quotes
		{`"abc'`, `"abc'`},
		{`'abc"`, `'abc"`},

		// Double quotes inside
		{`"he"llo"`, `he"llo`},
		{`"he\"llo"`, `he\"llo`},

		// Repeating quotes
		{`""`, ""},
		{`"""`, `"`},

		// Unicode and special characters
		{`"你好"`, "你好"},
		{`"😊"`, "😊"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := TrimQQ(tt.input)
			if result != tt.expected {
				t.Errorf("TrimQQ(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFlagValue(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flag     string
		expected string
	}{
		{"flag with space", []string{"cmd", "--config", "file.yaml"}, "config", "file.yaml"},
		{"flag with equals", []string{"cmd", "--config=file.yaml"}, "config", "file.yaml"},

		{"flag in middle", []string{"cmd", "--other", "x", "--config", "file.yaml", "--more", "y"}, "config", "file.yaml"},
		{"flag at end", []string{"cmd", "--other", "x", "--config", "file.yaml"}, "config", "file.yaml"},

		{"multiple flags, first wins", []string{"cmd", "--config", "first.yaml", "--config", "second.yaml"}, "config", "second.yaml"},
		{"multiple equals flags, first wins", []string{"cmd", "--config=first.yaml", "--config=second.yaml"}, "config", "second.yaml"},

		{"flag with empty value", []string{"cmd", "--config", ""}, "config", ""},
		{"flag with equals empty", []string{"cmd", "--config="}, "config", ""},

		{"flag with no value", []string{"cmd", "--config"}, "config", ""},
		{"flag not present", []string{"cmd", "--other", "value"}, "config", ""},
		{"flag prefix only", []string{"cmd", "--configfile"}, "config", ""},
		{"flag equals prefix only", []string{"cmd", "--configfile=value"}, "config", ""},

		{"flag similar name", []string{"cmd", "--configx", "value"}, "config", ""},
		{"flag similar equals", []string{"cmd", "--configx=value"}, "config", ""},

		{"empty args", []string{}, "config", ""},
		{"only program name", []string{"cmd"}, "config", ""},
		{"short flag ignored", []string{"cmd", "-config", "file.yaml"}, "config", ""},
		{"single dash with equals", []string{"cmd", "-config=file.yaml"}, "config", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlagValue(tt.args, tt.flag)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUints(t *testing.T) {
	tests := []struct {
		input    string
		expected []uint
	}{
		{"", nil},
		{"abc", nil},
		{"123", []uint{123}},
		{"abc123", []uint{123}},
		{"123abc", []uint{123}},
		{"abc123def", []uint{123}},
		{"abc123def456", []uint{123, 456}},
		{"0", []uint{0}},
		{"000123", []uint{123}},
		{"12 34 56", []uint{12, 34, 56}},
		{"12abc34def56", []uint{12, 34, 56}},
		{"1a2b3c", []uint{1, 2, 3}},
		{"1.2.3", []uint{1, 2, 3}},
		{"1,2,3", []uint{1, 2, 3}},
		{"  123  ", []uint{123}},
	}

	for _, test := range tests {
		result := Uints(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("uints(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestIsGitHash(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid hashes
		{"a1b2c3d", true}, // 7 chars
		{"0123456789abcdef0123456789abcdef01234567", true}, // 40 chars
		{"deadbeef", true},
		{"1234567", true},
		{"abcdef0", true},

		// Invalid due to length
		{"123456", false}, // too short
		{"0123456789abcdef0123456789abcdef012345678", false}, // 41 chars

		// Invalid characters
		{"g123456", false}, // 'g' is not hex
		{"12345z7", false}, // 'z' is not hex
		{"12345_7", false}, // '_' is not hex
		{"ABCDEF1", false}, // uppercase not allowed

		// Edge cases
		{"", false},
		{"ffffffffffffffffffffffffffffffffffffffff", true}, // all 'f's
		{"0000000", true}, // all '0's
	}

	for _, test := range tests {
		result := IsGitHash(test.input)
		if result != test.expected {
			t.Errorf("isGitHash(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestIsLessByNums(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool
	}{
		// Identical strings
		{"abc123", "abc123", false},

		// Git commit hash handling
		{"a1b2c3d", "abc123", true}, // a is a git hash
		{"abc123", "a1b2c3d", false},
		{"deadbeef", "1234567", true},
		{"123456G", "deadbeef", false},
		{"abcdef0", "abcdef0", false},

		// Numeric comparison
		{"file1", "file2", true},
		{"file2", "file1", false},
		{"file10", "file2", false},
		{"file2", "file10", true},
		{"v1.2.3", "v1.2.4", true},
		{"v1.2.10", "v1.2.4", false},
		{"v1.2.3", "v1.2.3", false},

		// Different number of numeric components
		{"v1.2", "v1.2.1", true},
		{"v1.2.1", "v1.2", false},
		{"v1.2.0", "v1.2", false}, // trailing zero

		// No numbers
		{"abc", "def", false},
		{"abc", "abc", false},

		// Mixed characters and numbers
		{"a1b2c3", "a1b2c4", true},
		{"a1b2c4", "a1b2c3", false},
		{"a1b2c3", "a1b2c3", false},

		// Leading zeros
		{"file001", "file1", false},
		{"file1", "file001", false},

		// Empty strings
		{"", "", false},
		{"", "abc123", true},
		{"abc123", "", false},

		// Git hash vs Git hash
		{"abcdef1", "abcdef2", true},
		{"abcde.2", "abcde.1", false},
		{"abcdef1", "abcdef1", false},

		// Git hash vs longer string with numbers
		{"abcdef1", "version123", true},
		{"version123", "abcdef1", false},

		// Long numeric sequences
		{"a1234567890", "a1234567891", true},
		{"v1234567891", "v1234567890", false},

		// Non-hex characters in non-hash strings
		{"abc123!", "abc123", false},
		{"abc123", "abc123!", false},
	}

	for _, test := range tests {
		result := IsLessByNums(test.a, test.b)
		if result != test.expected {
			t.Errorf("isLessByNums(%q, %q) = %v; want %v", test.a, test.b, result, test.expected)
		}
	}
}
