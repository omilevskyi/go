package datany

import (
	"reflect"
	"testing"

	. "github.com/omilevskyi/go/pkg/utils"
)

func TestLeafCount(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{
			name:     "Empty map",
			input:    map[string]any{},
			expected: 0,
		},
		{
			name: "Flat map",
			input: map[string]any{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: 3,
		},
		{
			name: "Nested map",
			input: map[string]any{
				"a": map[string]any{
					"b": 1,
					"c": 2,
				},
				"d": 3,
			},
			expected: 3,
		},
		{
			name: "Deeply nested map",
			input: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": 1,
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Mixed types",
			input: map[string]any{
				"a": 1,
				"b": map[string]any{
					"c": "hello",
					"d": map[string]any{
						"e": true,
					},
				},
			},
			expected: 3,
		},
		{
			name:     "Non-map input (leaf)",
			input:    42,
			expected: 1,
		},
		{
			name: "Map with nil values",
			input: map[string]any{
				"a": nil,
				"b": map[string]any{
					"c": nil,
				},
			},
			expected: 2,
		},
		{
			name: "Map with slices (should count slices as leaves)",
			input: map[string]any{
				"a": []int{1, 2, 3},
				"b": map[string]any{
					"c": []string{"x", "y"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LeafCount(tt.input)
			if result != tt.expected {
				t.Errorf("leafCount(%v) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAnyByPath(t *testing.T) {
	data := map[any]any{
		"service": map[any]any{
			"component": []any{
				map[any]any{
					"name": "auth",
					"port": 8080,
				},
				map[any]any{
					"name": "db",
					"port": 5432,
				},
			},
			"enabled": true,
		},
		"version": "1.0.0",
	}

	tests := []struct {
		name     string
		path     string
		expected any
		fullPath string
	}{
		{
			name:     "Top-level key",
			path:     "version",
			expected: "1.0.0",
			fullPath: "version",
		},
		{
			name:     "Nested map key",
			path:     "service.enabled",
			expected: true,
			fullPath: "service.enabled",
		},
		{
			name:     "Array element field",
			path:     "service.component[0].name",
			expected: "auth",
			fullPath: "service.component[0].name",
		},
		{
			name:     "Array element dotted field",
			path:     "service.component.[0].name",
			expected: "auth",
			fullPath: "service.component[0].name",
		},
		{
			name:     "Second array element field",
			path:     ".service.component[1].port",
			expected: 5432,
			fullPath: "service.component[1].port",
		},
		{
			name:     "Array element as root",
			path:     "service..component[1]",
			expected: map[any]any{"name": "db", "port": 5432},
			fullPath: "service.component[1]",
		},
		{
			name:     "Invalid path (nonexistent key)",
			path:     "service.unknown",
			expected: nil,
			fullPath: "",
		},
		{
			name:     "Invalid index",
			path:     "service.component[10].name",
			expected: nil,
			fullPath: "",
		},
		{
			name:     "Invalid index number",
			path:     "service.component[n].name",
			expected: nil,
			fullPath: "",
		},
		{
			name:     "Invalid format",
			path:     "service.component.name",
			expected: nil,
			fullPath: "",
		},
		{
			name:     "Empty path",
			path:     "",
			expected: data,
			fullPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, full := AnyByPath(data, tt.path)
			if !reflect.DeepEqual(val, tt.expected) || full != tt.fullPath {
				t.Errorf("anyByPath(%q) = (%v, %q); want (%v, %q)", tt.path, val, full, tt.expected, tt.fullPath)
			}
		})
	}
}

func TestSearchKeys(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		predicate func(string) bool
		expected  []string
	}{
		{
			name: "Flat map, match all",
			data: map[any]any{
				"a": 1,
				"b": 2,
			},
			predicate: func(s string) bool { return true },
			expected:  []string{".a", ".b"},
		},
		{
			name: "Flat map, match none",
			data: map[any]any{
				"a": 1,
				"b": 2,
			},
			predicate: func(s string) bool { return false },
			expected:  nil,
		},
		{
			name: "Nested map, match some",
			data: map[any]any{
				"outer": map[any]any{
					"inner": 42,
					"skip":  true,
				},
			},
			predicate: func(s string) bool { return s == "inner" },
			expected:  []string{".outer.inner"},
		},
		{
			name: "Map with slice, match key",
			data: map[any]any{
				"list": []any{
					map[any]any{"name": "item1"},
					map[any]any{"name": "item2"},
				},
			},
			predicate: func(s string) bool { return s == "name" },
			expected:  []string{".list[0].name", ".list[1].name"},
		},
		{
			name: "Deep nesting with mixed types",
			data: map[any]any{
				"root": map[any]any{
					"child": []any{
						map[any]any{"target": 1},
						map[any]any{"other": 2},
					},
				},
			},
			predicate: func(s string) bool { return s == "target" },
			expected:  []string{".root.child[0].target"},
		},
		{
			name:      "Empty map",
			data:      map[any]any{},
			predicate: func(s string) bool { return true },
			expected:  nil,
		},
		{
			name:      "Nil input",
			data:      nil,
			predicate: func(s string) bool { return true },
			expected:  nil,
		},
		{
			name:      "Non-map/slice input",
			data:      123,
			predicate: func(s string) bool { return true },
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Arrange(SearchKeys(tt.data, tt.predicate))
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("searchKeys(...) = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestKeyAndIndex(t *testing.T) {
	tests := []struct {
		input         string
		expectedKey   string
		expectedIndex int
	}{
		{"key[0]", "key", 0},
		{"key[123]", "key", 123},
		{"key[-1]", "key", -1},
		{"key[999999]", "key", 999999},
		{"key[abc]", "", -1},
		{"key[", "", -1},
		{"key]", "", -1},
		{"key", "", -1},
		{"[0]", "", 0},
		{"", "", -1},
		{"key[[2]", "key[", 2},
		{"key[0][1]", "key", 0},    // ?
		{"key[0]extra", "key", 0},  // ?
		{"key[0][1][2]", "key", 0}, // ?
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, index := keyAndIndex(tt.input)
			if key != tt.expectedKey || index != tt.expectedIndex {
				t.Errorf("keyAndIndex(%q) = (%q, %d), want (%q, %d)", tt.input, key, index, tt.expectedKey, tt.expectedIndex)
			}
		})
	}
}

func TestStringByPath(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		path     string
		expected string
	}{
		{
			name: "Nested map with valid path",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": "value",
					},
				},
			},
			path:     ".level1.level2.key",
			expected: "value",
		},
		{
			name: "Nested map with invalid path",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": "value",
					},
				},
			},
			path:     "level1.level2.invalid",
			expected: "",
		},
		{
			name: "Nested map with empty path",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": "value",
					},
				},
			},
			path:     "",
			expected: "",
		},
		{
			name: "Nested map with missing key",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": "value",
					},
				},
			},
			path:     "level1.level2.missing",
			expected: "",
		},
		{
			name: "Slice with valid index",
			data: []any{
				map[any]any{"key": "value"},
				map[any]any{"key": "another value"},
			},
			path:     "[1].key",
			expected: "another value",
		},
		{
			name: "Slice with invalid index",
			data: []any{
				map[any]any{"key": "value"},
				map[any]any{"key": "another value"},
			},
			path:     "[2].key",
			expected: "",
		},
		{
			name: "Slice with empty path",
			data: []any{
				map[any]any{"key": "value"},
				map[any]any{"key": "another value"},
			},
			path:     "",
			expected: "",
		},
		{
			name: "Slice with missing key",
			data: []any{
				map[any]any{"key": "value"},
				map[any]any{"key": "another value"},
			},
			path:     "[1].missing",
			expected: "",
		},
		{
			name: "Numeric value",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": 123,
					},
				},
			},
			path:     "level1.level2.key",
			expected: "123",
		},
		{
			name: "Nil value",
			data: map[any]any{
				"level1": map[any]any{
					"level2": map[any]any{
						"key": nil,
					},
				},
			},
			path:     "level1.level2.key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringByPath(tt.data, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStringByPath2(t *testing.T) {
	type testCase struct {
		name     string
		data     any
		path     string
		expected string
	}

	tests := []testCase{
		// Scalar types
		{"Bool true", map[any]any{"a": true}, "a", "true"},
		{"Bool false", map[any]any{"a": false}, "a", "false"},
		{"Int", map[any]any{"a": 42}, "a", "42"},
		{"Float", map[any]any{"a": 3.14}, "a", "3.14"},
		{"String", map[any]any{"a": "hello"}, "a", "hello"},

		// Nested structures
		{"Nested map", map[any]any{"a": map[any]any{"b": 123}}, "a.b", "123"},
		{"Deep nesting", map[any]any{"a": map[any]any{"b": map[any]any{"c": "deep"}}}, "a.b.c", "deep"},

		// Slices
		{"List index", map[any]any{"a": []any{"zero", "one"}}, "a[1]", "one"},
		{"List of maps", map[any]any{"a": []any{map[any]any{"b": "val"}}}, "a[0].b", "val"},

		// Edge cases
		{"Empty path", map[any]any{"a": "x"}, "", ""},
		{"Nonexistent path", map[any]any{"a": "x"}, "b", ""},
		{"Nil data", nil, "a", ""},
		{"Nil value", map[any]any{"a": nil}, "a", ""},

		// Unsupported types
		{"Unsupported type: struct", map[any]any{"a": struct{}{}}, "a", ""},
		{"Unsupported type: func", map[any]any{"a": func() {}}, "a", ""},
		{"Unsupported type: chan", map[any]any{"a": make(chan int)}, "a", ""},
		{"Unsupported type: complex", map[any]any{"a": complex(1, 2)}, "a", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := StringByPath(tc.data, tc.path)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}
