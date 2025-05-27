package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const nl = "\n"

// Quote character for TrimQQ() and EnQQ()
var Quote byte = '"'

// IsErr checks if an error is non-nil, logs it with source location and optional context,
// and optionally exits the program with the given return code.
// filename.go:line optionalFunctionName(): optionalMessage: error
func IsErr(err error, rc int, slice ...string) bool {
	if err == nil {
		return false
	}
	pc, msg, fname := make([]uintptr, 15), strings.Join(slice, ""), ""
	f, _ := runtime.CallersFrames(pc[:runtime.Callers(2, pc)]).Next()
	if len(slice) < 1 && f.Function != "" {
		fname = " " + f.Function[strings.LastIndex(f.Function, ".")+1:] + "()"
	}
	if msg != "" {
		msg += ": "
	}
	_, _ = fmt.Fprint(os.Stderr, filepath.Base(f.File), ":", f.Line, fname, ": ", msg, err, nl)
	if rc > 0 {
		os.Exit(rc)
	}
	return true
}

// CallSite returns the file name and line number of the caller for debugging or logging purposes.
func CallSite() string {
	pc := make([]uintptr, 15)
	f, _ := runtime.CallersFrames(pc[:runtime.Callers(2, pc)]).Next()
	s := filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
	if f.Function != "" {
		s += " " + f.Function[strings.LastIndex(f.Function, ".")+1:] + "()"
	}
	return s
}

// Fringerr wraps an error with file name, line number, and function name.
func Fringerr(e error) error {
	pc := make([]uintptr, 15)
	f, _ := runtime.CallersFrames(pc[:runtime.Callers(2, pc)]).Next()
	s := filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
	if f.Function != "" {
		s += " " + f.Function[strings.LastIndex(f.Function, ".")+1:] + "()"
	}
	return fmt.Errorf("%s: %w", s, e)
}

// Keys returns a slice containing all the keys from the given map.
func Keys[K comparable, V any](m map[K]V) []K {
	if i := len(m); i > 0 {
		slice := make([]K, i)
		for k := range m {
			i--
			slice[i] = k
		}
		return slice
	}
	return nil
}

// Arrange sorts a slice of strings in ascending order; the returned slice is the same as the input.
func Arrange(s []string) []string {
	if len(s) > 0 {
		sort.Strings(s)
		return s
	}
	return nil
}

// ArrangeCopy sorts a slice of strings in ascending order; the returned slice is a new slice.
func ArrangeCopy(s []string) []string {
	if slen := len(s); slen > 0 {
		result := make([]string, slen)
		copy(result, s)
		sort.Strings(result)
		return result
	}
	return nil
}

// ArrangB sorts a slice of []byte in ascending order; the returned slice is the same as the input.
func ArrangB(s [][]byte) [][]byte {
	if len(s) > 0 {
		sort.Slice(s, func(i, j int) bool {
			return bytes.Compare(s[i], s[j]) < 0
		})
		return s
	}
	return nil
}

// ArrangBCopy sorts a slice of []byte in ascending order; the returned slice is a new slice.
func ArrangBCopy(s [][]byte) [][]byte {
	if slen := len(s); slen > 0 {
		result := make([][]byte, slen)
		copy(result, s)
		sort.Slice(result, func(i, j int) bool {
			return bytes.Compare(s[i], s[j]) < 0
		})
		return result
	}
	return nil
}

// An [unsuccessful] attempt to implement Arrange() on generics
// func ArrangeT[T comparable](slice []T) []T {
// 	switch v := any(slice).(type) {
// 	case []string:
// 		sort.Strings(v)
// 	case []int:
// 		sort.Ints(v)
// 	default:
// 		panic("Unsupported type")
// 	}
// 	return slice
// }

// Compact returns the passed slice with all empty strings removed.
func Compact(slice []string) []string {
	if len(slice) > 0 {
		deleteItems(&slice, "")
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}

// It seems to be faster, but not safer.
// P.S. There was also some doubt about cutting a slice passed by value to the Compact() function.
//
// --- FAIL: TestNewAlias (0.00s)
//
//	--- FAIL: TestNewAlias/Nested_keys_with_comments (0.00s)
//		alias_test.go:533:
//			AltNewAlias() = &{map[key1:{                    }] map[key1:comment 1 key1+key2:comment 2 key1+key2+key2+key3:comment 3]}, <nil>,
//		             want &{map[key1:{                    }] map[key1:comment 1 key1+key2:comment 2 key1+key2+key3:comment 3]}, <nil>
func deleteItems[T comparable](slicePtr *[]T, targets ...T) {
	if slicePtr != nil {
		s, slen, j := *slicePtr, len(*slicePtr), 0
		for i := 0; i < slen; i++ {
			if !slices.Contains(targets, s[i]) {
				s[j] = s[i]
				j++
			}
		}
		*slicePtr = s[:j] // It is better to add 'if j == 0 { *slicePtr = nil }', but deleteItems() is private and is used by Compact() only.
	}
}

// CompactCopy returns a new slice with all empty strings removed.
func CompactCopy(slice []string) []string {
	if len(slice) > 0 {
		return deleteItemsCopy(slice, "")
	}
	return nil
}

// deleteItemsCopy returns a new slice with specified target elements removed.
func deleteItemsCopy[T comparable](slice []T, targets ...T) []T {
	if slen := len(slice); slen > 0 {
		j := 0
		for i := 0; i < slen; i++ {
			if !slices.Contains(targets, slice[i]) {
				j++
			}
		}
		if j > 0 {
			result := make([]T, j)
			for i := slen - 1; i >= 0; i-- {
				if !slices.Contains(targets, slice[i]) {
					j--
					result[j] = slice[i]
				}
			}
			return result
		}
	}
	return nil
}

// Distinct returns the passed slice with duplicate elements removed, preserving the original order.
func Distinct[T comparable](slice []T) []T {
	if slen := len(slice); slen > 0 {
		seen, j := make(map[T]struct{}, slen), 0
		for i := 0; i < slen; i++ {
			if _, ok := seen[slice[i]]; !ok {
				seen[slice[i]] = struct{}{}
				if j < i {
					slice[j] = slice[i]
				}
				j++
			}
		}
		if j > 0 {
			return slice[:j]
		}
	}
	return []T(nil)
}

// DistinctCopy returns a new slice with duplicate elements removed, preserving the original order.
func DistinctCopy[T comparable](slice []T) []T {
	if slen := len(slice); slen > 0 {
		result, seen, j := make([]T, slen), make(map[T]struct{}, slen), 0
		for i := 0; i < slen; i++ {
			if _, ok := seen[slice[i]]; !ok {
				seen[slice[i]] = struct{}{}
				result[j] = slice[i]
				j++
			}
		}
		if j > 0 {
			return result[:j]
		}
	}
	return nil
}

// TrimExt removes the file extension from a filename.
func TrimExt(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// TrimQQ removes surrounding quote characters from a string if present.
func TrimQQ(s string) string {
	if slen := len(s); slen > 1 && s[0] == Quote && s[slen-1] == Quote {
		return s[1 : slen-1]
	}
	return s
}

// EnQQ encloses a string in quote characters.
func EnQQ(s string) string {
	return string(Quote) + s + string(Quote)
}

// IntToLetter maps an integer to a letter (a–z, A–Z) in a cyclic manner.
func IntToLetter(num int) rune {
	syms := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	}
	return rune(syms[num%len(syms)])
}
