package datany

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	// PathSeparator - path component separator for YAML
	PathSeparator byte = '.'

	// IndexOpen - opening delimiter of index within list
	IndexOpen byte = '['

	// IndexClose - closing delimiter of index within list
	IndexClose byte = ']'
)

// LeafCount returns the number of non-map (leaf) values in a nested map structure.
func LeafCount(m any) (count int) {
	var countHelper func(any)
	countHelper = func(m any) {
		if v := reflect.ValueOf(m); v.Kind() == reflect.Map {
			for _, k := range v.MapKeys() {
				countHelper(v.MapIndex(k).Interface())
			}
			return
		}
		count++
	}
	countHelper(m)
	return count
}

// StringByPath retrieves a value from a nested structure by path and returns it as a string.
func StringByPath(data any, path string) string {
	if value, _ := AnyByPath(data, path); value != nil {
		switch v := value.(type) {
		case bool, int, float64:
			return fmt.Sprint(v)
		case string:
			return v
		}
		// return fmt.Sprintf("<%T>", value)
	}
	return ""
}

// SearchKeys traverses a nested map or slice structure and returns all key paths that match a given predicate.
func SearchKeys(data any, f func(string) bool) []string {
	prefix, results := "", []string(nil)
	var searchKeysRecursive func(any, string)
	searchKeysRecursive = func(data any, prefix string) {
		switch dt := data.(type) {
		case map[any]any:
			for k, v := range dt {
				keyStr := fmt.Sprint(k)
				newPrefix := prefix + string(PathSeparator) + keyStr
				if f(keyStr) {
					results = append(results, newPrefix)
				}
				searchKeysRecursive(v, newPrefix)
			}
		case []any:
			for i, v := range dt {
				searchKeysRecursive(v, fmt.Sprintf("%s[%d]", prefix, i))
			}
		}
	}
	searchKeysRecursive(data, prefix)
	return results
}

// SplitOnTokens splits YAML path on string tokens, supporting nested brackets
func SplitOnTokens(input string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		start, bracketDepth := -1, 0

		for i := 0; i < len(input); i++ {
			if bracketDepth > 0 {
				switch input[i] {
				case IndexOpen:
					bracketDepth++
				case IndexClose:
					bracketDepth--
					if bracketDepth == 0 {
						out <- input[start : i+1]
						start = -1
					}
				}
				continue
			}

			switch input[i] {
			case PathSeparator:
				if start != -1 {
					out <- input[start:i]
					start = -1
				}
			case IndexOpen:
				if start != -1 {
					out <- input[start:i]
				}
				start, bracketDepth = i, 1
			default:
				if start == -1 {
					start = i
				}
			}
		}

		if start != -1 {
			out <- input[start:]
		}
	}()

	return out
}

// AnyByPath navigates a nested map or slice structure using a dot-separated path and returns the value and resolved path.
func AnyByPath(data any, path string) (any, string) {
	builder, key, delim, i, err, ok := strings.Builder{}, "", "", 0, error(nil), false
	builder.Grow(len(path))
	for key = range SplitOnTokens(path) {
		switch dt := data.(type) {
		case map[any]any:
			if data, ok = dt[key]; ok { // service.component.name
				_, _ = builder.WriteString(delim)
				_, _ = builder.WriteString(key)
				delim = string(PathSeparator)
				continue
			}
		case []any: // service.component.[0]
			if i = len(key) - 1; i > 1 && key[0] == IndexOpen && key[i] == IndexClose {
				if i, err = strconv.Atoi(key[1:i]); err == nil && 0 <= i && i < len(dt) {
					_, _ = builder.WriteString(delim)
					_ = builder.WriteByte(IndexOpen)
					_, _ = builder.WriteString(strconv.Itoa(i))
					_ = builder.WriteByte(IndexClose)
					data, delim = dt[i], string(PathSeparator)
					continue
				}
			}
		}
		return nil, ""
	}
	return data, builder.String()
}
