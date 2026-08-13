package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	e2pSep    = "/"
	e2pStLet  = 'h'
	e2pFinLet = 'z'
	e2pStDgt  = '1'
	e2pFinDgt = '9'

	epSep = ':'
	sp    = ' '
	tab   = '\t'
	lf    = '\n'
)

var (
	loadConfig   = config.LoadDefaultConfig
	newElbClient = elasticloadbalancingv2.NewFromConfig
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

func printEnvs(w io.Writer, items []itemT) error {
	const envStr = " environment"

	n := len(items)
	if n == 0 {
		return nil
	}

	size := len("999") + len(envStr) + 2 + 1 // "999" + " environment" + "s:" + '\n'
	for _, item := range items {
		size += 1 + len(item.env) + 1 + len(item.profile) // leading space + env
	}

	var b strings.Builder
	b.Grow(size)

	if n == 1 {
		b.WriteByte('1')
		b.WriteString(envStr)
	} else {
		b.WriteString(strconv.Itoa(n))
		b.WriteString(envStr)
		b.WriteByte('s')
	}
	b.WriteByte(':')

	for _, item := range items {
		b.WriteByte(' ')
		b.WriteString(item.env)
		b.WriteByte(epSep)
		b.WriteString(item.profile)
	}

	b.WriteByte(lf)

	_, err := io.WriteString(w, b.String())
	return err
}

// envProfile splits a string in the form "env:profile" and returns the first
// two components. If no profile is specified, the second return value is empty.
// Any additional colon-separated components are ignored.
func envProfile(s string) (string, string) {
	// 	parts := strings.SplitN(s, string(envPrfSep), 3)
	// 	if len(parts) == 1 {
	// 		return strings.TrimSpace(parts[0]), ""
	// 	}
	// 	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	// i := strings.IndexRune(s, envProfSep)
	// if i < 0 {
	// 	return strings.TrimSpace(s), ""
	// }
	// j := strings.IndexRune(s[i+1:], envProfSep)
	// if j < 0 {
	// 	j = len(s) - i - 1
	// }
	// return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1 : i+1+j])

	i, n := 0, len(s)
	e0 := 0                                       // environment start
	for e0 < n && (s[e0] == sp || s[e0] == tab) { // trim leading spaces from environment
		e0++
	}
	e1, p0, p1 := e0, n, n // environment finish, profile start, profile finish
	for i = e0; i < n; i++ {
		if s[i] == epSep {
			e1 = i     // environment finish
			p0 = i + 1 // profile start
			break
		}
	}
	if e1 == e0 && e1 < n && s[e1] != epSep { // no profile
		e1 = n
	}
	for e1 > e0 && (s[e1-1] == sp || s[e1-1] == tab) { // trim trailing spaces from environment
		e1--
	}
	if p0 < n {
		for p0 < n && (s[p0] == sp || s[p0] == tab) { // trim leading spaces from profile
			p0++
		}
		for i = p0; i < n; i++ {
			if s[i] == epSep {
				p1 = i // profile finish
				break
			}
		}
		for p1 > p0 && (s[p1-1] == sp || s[p1-1] == tab) { // trim trailing spaces from profile
			p1--
		}
	}

	return s[e0:e1], s[p0:p1]
}

func printProfiles(w io.Writer, profiles []ProfileT) error {
	var b strings.Builder

	for _, p := range profiles {
		b.WriteString(p.Name)
		b.WriteByte('(')
		if p.envs == nil {
			b.WriteByte('0')
		} else {
			b.WriteString(strconv.Itoa(len(*p.envs)))
		}
		b.WriteString("):")
		b.WriteByte(lf)
		for _, e := range *p.envs {
			b.WriteByte(sp)
			b.WriteString(e.Name)
			b.WriteByte('(')
			if e.LBs == nil {
				b.WriteByte('0')
			} else {
				b.WriteString(strconv.Itoa(len(*e.LBs)))
			}
			b.WriteByte(')')
		}
		b.WriteByte(lf)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func verbf(lvl verbosityLevelT, format string, a ...any) error {
	if lvl <= verbosityLevel && len(format) > 0 {
		if format[len(format)-1] != lf {
			format += string(lf)
		}
		_, err := fmt.Fprintf(os.Stderr, format, a...)
		if err != nil {
			return err
		}
	}
	return nil
}

func printResult(w io.Writer, profiles []ProfileT) error {
	var err error
	for _, p := range profiles {
		if p.envs == nil {
			continue
		}
		for _, e := range *p.envs {
			if e.LBs == nil {
				continue
			}
			for _, lb := range e.SortLBs() {
				if lb.TgARNs == nil || len(*lb.TgARNs) < 1 {
					continue
				}
				if _, err = fmt.Fprintln(w, lb.ARN); err != nil {
					return err
				}
				for _, tg := range ut.Arrange(*lb.TgARNs) {
					if _, err = fmt.Fprintln(w, " ", tg); err != nil {
						return err
					}
				}
				if _, err = fmt.Fprintln(w); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
