package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
	ut "github.com/omilevskyi/go/pkg/utils"
	// "github.com/davecgh/go-spew/spew"
)

const (
	appName    = "unhealthy-tgs"
	timeoutSec = 60
)

var (
	version = "<dev>" // -ldflags -X main.version=v0.0.0 -X main.commit=[[:xdigit:]]+
	commit  = "<none>"
)

type verbosityLevelT uint8

var verbosityLevel verbosityLevelT

const (
	// vEmerg   verbosityLevelT = iota // 0 - system is unusable
	// vAlert                          // 1 - action must be taken immediately
	// vCrit                           // 2 - critical conditions
	// vErr                            // 3 - error conditions
	// vWarning                        // 4 - warning conditions
	// vNotice                         // 5 - normal but significant condition
	// vInfo                           // 6 - informational
	// vDebug                          // 7 - debug-level messages
	vError  verbosityLevelT = iota // error conditions
	vNotice                        // normal but significant condition
	vInfo                          // informational
	vDebug                         // debug-level messages
)

func main() {
	start := time.Now()

	var isHelp, isVersion bool
	var vrbLvl int

	flag.BoolVar(&isHelp, "help", false, "Show usage message")
	flag.BoolVar(&isVersion, "version", false, "Show version information")
	flag.IntVar(&vrbLvl, "verbose", 0, "Enable verbose output with specified level")
	flag.Parse()

	if isHelp {
		fmt.Fprintln(os.Stderr, "Usage:", appName, "[-help] [-version] [-verbose] [environments...]")
		os.Exit(0)
	}

	if isVersion {
		fmt.Fprintln(os.Stderr, "Version: "+version+", Commit: "+commit)
		os.Exit(0)
	}

	verbosityLevel = verbosityLevelT(vrbLvl)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	var err error
	var routines []ProfileT
	if args := flag.Args(); isatty.IsTerminal(os.Stdin.Fd()) && len(args) > 0 {
		var items []itemT
		items, routines, err = buildProfiles(ctx, args)
		ut.IsErr(err, 201)

		_ = verbosityLevel > 0 && printEnvs(os.Stderr, items) == nil

		err = groupLoadBalancers(ctx, routines)
		ut.IsErr(err, 202, "groupLoadBalancers()")
		_ = verbosityLevel > 1 && printProfiles(os.Stderr, routines) == nil
	} else {
		var dummyEnv string
		dummyEnv, routines, err = buildSingleProfile(ctx, "AWS_PROFILE")
		ut.IsErr(err, 201, "buildSingleProfile()")

		err = readArns(os.Stdin, routines, dummyEnv)
		ut.IsErr(err, 202, "readArns()")
		verbf(vNotice, "load balancers: %d", lbCount(routines))
	}

	err = selectUnhealthy(ctx, routines)
	ut.IsErr(err, 203, "selectUnhealthy()")

	err = printResult(os.Stdout, routines)
	ut.IsErr(err, 204, "printResult()")

	verbf(vNotice, "Time spent: %.1f seconds", time.Since(start).Seconds())
	// spew.Dump(rc)
}
