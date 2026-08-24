package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	txt "github.com/omilevskyi/go/pkg/text"
	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	e2pSep    = "/"
	e2pStLet  = 'h'
	e2pFinLet = 'z'
	e2pStDgt  = '1'
	e2pFinDgt = '9'

	epSep = ':'
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
func readArns(r io.Reader, profiles []ProfileT, envName string) error {
	var env *EnvT
	for _, p := range profiles {
		for i := range *p.envs {
			if (*p.envs)[i].Name == envName {
				env = &(*p.envs)[i]
				break
			}
		}
	}
	if env == nil {
		return errors.New("environment name is not found")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	t := txt.New(
		txt.WithSkipWhiteSpace(true), txt.WithTrimLeadSpace(true), txt.WithTrimTrailSpace(true),
	)

	if env.LBs == nil {
		env.LBs = new(make([]LbT, 0, t.LineCount(data)))
	}

	for line := range t.Lines(data) {
		*env.LBs = append(*env.LBs, LbT{ARN: string(line), TgARNs: new(make([]string, 0)), mu: new(sync.Mutex)})
	}

	return nil
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
		if item.profile != "" {
			b.WriteByte(epSep)
			b.WriteString(item.profile)
		}
	}

	b.WriteByte(txt.LF)

	_, err := io.WriteString(w, b.String())
	return err
}

// envProfile splits a string in the form "env:profile" and returns the first
// two components. If no profile is specified, the second return value is empty.
// Any additional colon-separated components are ignored.
func envProfile(s string) (string, string) {
	i, n := 0, len(s)
	e0 := 0                                              // environment start
	for e0 < n && (s[e0] == txt.SP || s[e0] == txt.HT) { // trim leading spaces from environment
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
	for e1 > e0 && (s[e1-1] == txt.SP || s[e1-1] == txt.HT) { // trim trailing spaces from environment
		e1--
	}
	if p0 < n {
		for p0 < n && (s[p0] == txt.SP || s[p0] == txt.HT) { // trim leading spaces from profile
			p0++
		}
		for i = p0; i < n; i++ {
			if s[i] == epSep {
				p1 = i // profile finish
				break
			}
		}
		for p1 > p0 && (s[p1-1] == txt.SP || s[p1-1] == txt.HT) { // trim trailing spaces from profile
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
		b.WriteByte(txt.LF)
		for _, e := range *p.envs {
			b.WriteByte(txt.SP)
			b.WriteString(e.Name)
			b.WriteByte('(')
			if e.LBs == nil {
				b.WriteByte('0')
			} else {
				b.WriteString(strconv.Itoa(len(*e.LBs)))
			}
			b.WriteByte(')')
		}
		b.WriteByte(txt.LF)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func verbf(lvl verbosityLevelT, format string, a ...any) {
	if lvl <= verbosityLevel && len(format) > 0 {
		if format[len(format)-1] != txt.LF {
			format += string(txt.LF)
		}
		fmt.Fprintf(os.Stderr, format, a...)
	}
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

// lbCount returns the number of load balancers associated with the idxs[1]
// environment of the idxs[0] profile.
//
// Returns 0 if any required level of the profile/environment/load balancer
// hierarchy is missing or invalid.
//
// The implementation intentionally favors readability over a more compact form.
func lbCount(profiles []ProfileT, idxs ...int) int {
	i := 0
	if len(idxs) > 0 {
		i = idxs[0]
	}
	if len(profiles) == 0 || i >= len(profiles) || profiles[i].envs == nil {
		return 0
	}

	envs := *profiles[i].envs
	if len(idxs) > 1 {
		i = idxs[1]
	} else {
		i = 0
	}
	if len(envs) == 0 || i >= len(envs) || envs[i].LBs == nil {
		return 0
	}

	return len(*envs[i].LBs)
}
