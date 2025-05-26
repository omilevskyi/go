package datany

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	// PathDelim - path component delimiter for YAML
	PathDelim byte = '.'

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
	return string([]byte(nil))
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
				newPrefix := prefix + string(PathDelim) + keyStr
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

// keyAndIndex extracts a key and index from a string formatted like "key[0]".
func keyAndIndex(s string) (string, int) {
	// Old implementation which goes from tail
	// if slen := len(s); slen > 1 && s[slen-1] == IndexClose {
	// 	for i := slen - 2; i >= 0; i-- {
	// 		if s[i] == IndexOpen {
	// 			if n, err := strconv.Atoi(string(s[i+1 : slen-1])); err == nil {
	// 				return s[:i], n
	// 			}
	// 			break
	// 		}
	// 	}
	// }
	// if i := strings.IndexByte(s, IndexOpen); i >= 0 {
	// 	if j := strings.IndexByte(s[i+1:], IndexClose); j >= 0 {
	// 		if n, err := strconv.Atoi(string(s[i+1 : i+j+1])); err == nil {
	// 			return s[:i], n
	// 		}
	// 	}
	// }
	// if j := strings.IndexByte(s[i+1:], IndexClose); j >= 0 {
	// 	if n, err := strconv.Atoi(string(s[i+1 : i+j+1])); err == nil {
	// 		return s[:i], n
	// 	}
	// }
	if i := strings.IndexByte(s, IndexOpen); i >= 0 {
		for j := i + 1; j < len(s); j++ {
			switch s[j] {
			case IndexOpen:
				i = j
			case IndexClose:
				if n, err := strconv.Atoi(string(s[i+1 : j])); err == nil {
					return s[:i], n
				}
			}
		}
	}
	return string([]byte(nil)), -1
}

// AnyByPath navigates a nested map or slice structure using a dot-separated path and returns the value and resolved path.
func AnyByPath(data any, path string) (any, string) {
	builder, key, delim, i, slice, ok := strings.Builder{}, "", "", 0, []any(nil), false
	builder.Grow(len(path))
	for key = range strings.SplitSeq(path, string(PathDelim)) {
		if key == "" {
			continue
		}
		switch dt := data.(type) {
		case map[any]any:
			if data, ok = dt[key]; ok { // service.component.name
				_, _ = builder.WriteString(delim + key)
				delim = string(PathDelim)
				continue
			}
			if key, i = keyAndIndex(key); 0 <= i { // service.component[0].name
				if data, ok = dt[key]; ok {
					_, _ = builder.WriteString(delim + key)
					delim = string(PathDelim)
					if slice, ok = data.([]any); ok && 0 <= i && i < len(slice) {
						_, _ = builder.WriteString(fmt.Sprintf("%c%d%c", IndexOpen, i, IndexClose))
						data = slice[i]
						continue
					}
				}
			}
		case []any: // service.component.[0].name -> service.component[0].name
			// fmt.Fprintln(os.Stderr, "[]any", key)
			if _, i = keyAndIndex(key); 0 <= i && i < len(dt) {
				_, _ = builder.WriteString(fmt.Sprintf("%c%d%c", IndexOpen, i, IndexClose))
				data = dt[i]
				continue
			}
		}
		return nil, string([]byte(nil))
	}
	return data, builder.String()
}
