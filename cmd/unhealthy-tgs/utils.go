package main

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

const (
	e2pSep    = "/"
	e2pStLet  = 'h'
	e2pFinLet = 'z'
	e2pStDgt  = '1'
	e2pFinDgt = '9'
)

// env2path converts environment names with an h-z suffix letter followed by a
// 1-9 digit into a hierarchical path.
// For example, "headh2" becomes "/head/h/2/".
// All other values are returned as a simple "/<env>/" path.
func env2path(s string) string {
	if n := len(s); n > 2 {
		prev, last := s[n-2], s[n-1]
		if e2pStDgt <= last && last <= e2pFinDgt && e2pStLet <= prev && prev <= e2pFinLet {
			return e2pSep + s[:n-2] + e2pSep + string(prev) + e2pSep + string(last) + e2pSep
		}
	}

	return e2pSep + s + e2pSep
}

// readArns reads non-empty, whitespace-trimmed lines from r and returns them
// as a slice. Empty lines are ignored.
func readArns(r io.Reader) ([]string, error) {
	var arns []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := bytes.TrimSpace(scanner.Bytes()); len(line) > 0 { // line is valid until next scanner.Scan() call
			arns = append(arns, string(line))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(arns) > 0 {
		return arns, nil
	}
	return nil, nil
}

func printEnvs(w io.Writer, envs []string) error {
	const envsStr = " environments:"

	n := len(envs)
	if n == 0 {
		return nil
	}

	size := 4 + len(envsStr) + 1 // '\n'
	for _, env := range envs {
		size += 1 + len(env) // leading space + env
	}

	var b strings.Builder
	b.Grow(size)

	if n == 1 {
		b.WriteString("1 environment:")
	} else {
		b.WriteString(strconv.Itoa(n))
		b.WriteString(envsStr)
	}

	for _, env := range envs {
		b.WriteByte(' ')
		b.WriteString(env)
	}

	b.WriteByte('\n')

	_, err := io.WriteString(w, b.String())
	return err
}
