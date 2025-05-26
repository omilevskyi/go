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

// Quote character for TrimQQ() and EnQQ()
var Quote byte = '"'

// IsErr checks if an error is non-nil, logs it with source location and optional context,
// and optionally exits the program with the given return code.
func IsErr(err error, rc int, slice ...string) bool {
	if err == nil {
		return false
	}
	pc, msg := make([]uintptr, 15), strings.Join(slice, "")
	f, _ := runtime.CallersFrames(pc[:runtime.Callers(2, pc)]).Next()
	if msg == "" {
		msg = "ERROR"
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s:%d %s: %v\n", filepath.Base(f.File), f.Line, msg, err)
	if rc > 0 {
		os.Exit(rc)
	}
	return true
}

// CallSite returns the file name and line number of the caller for debugging or logging purposes.
func CallSite() string {
	pc := make([]uintptr, 15)
	f, _ := runtime.CallersFrames(pc[:runtime.Callers(2, pc)]).Next()
	return filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
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

// ArrangeNew sorts a slice of strings in ascending order; the returned slice is a new slice.
func ArrangeNew(s []string) []string {
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

// ArrangBNew sorts a slice of []byte in ascending order; the returned slice is a new slice.
func ArrangBNew(s [][]byte) [][]byte {
	if slen := len(s); slen > 0 {
		result := make([][]byte, slen)
		copy(result, s)
		sort.Slice(result, func(i, j int) bool {
			return bytes.Compare(s[i], s[j]) < 0
		})
		return s
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

// Compact returns a new slice with all empty strings removed.
func Compact(slice []string) []string {
	// return deleteItems(slice, "")
	if len(slice) > 0 {
		deleteItemsInPlace(&slice, "")
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}

// deleteItems returns a new slice with specified target elements removed.
func deleteItems[T comparable](slice []T, targets ...T) []T { // nolint:unused
	slen := len(slice)
	if slen < 1 {
		return nil
	}
	j := 0
	for i := 0; i < slen; i++ {
		if !slices.Contains(targets, slice[i]) {
			j++
		}
	}
	result := make([]T, j)
	for i := slen - 1; i >= 0; i-- {
		if !slices.Contains(targets, slice[i]) {
			j--
			result[j] = slice[i]
		}
	}
	return result
}

// The previous implementation of the function is provided as a reference example
// It seems to be faster, but not safer.
// P.S. There was also some doubt about updating a slice passed by value to the Compact() function.
func deleteItemsInPlace[T comparable](slicePtr *[]T, targets ...T) { // nolint:unused
	if slicePtr != nil {
		s, slen, j := *slicePtr, len(*slicePtr), 0
		for i := 0; i < slen; i++ {
			if !slices.Contains(targets, s[i]) {
				s[j] = s[i]
				j++
			}
		}
		*slicePtr = s[:j]
	}
}

// Distinct returns a new slice with duplicate elements removed, preserving the original order.
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
		return slice[:j]
	}
	return []T(nil)
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
