package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/mattn/go-isatty"
	ut "github.com/omilevskyi/go/pkg/utils"
	// "github.com/davecgh/go-spew/spew"
)

const appName = "unhealthy-tgs"

var (
	version = "<dev>" // -ldflags -X main.version=v0.0.0 -X main.commit=[[:xdigit:]]+
	commit  = "<none>"
)

func main() {
	start := time.Now()

	var isHelp, isVerbose, isVersion bool

	flag.BoolVar(&isHelp, "help", false, "Show usage message")
	flag.BoolVar(&isVersion, "version", false, "Show version information")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if isHelp {
		fmt.Fprintln(os.Stderr, "Usage:", appName, "[-help] [-version] [-verbose] [environments...]")
		os.Exit(0)
	}

	if isVersion {
		fmt.Fprintln(os.Stderr, "Version: "+version+", Commit: "+commit)
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	ut.IsErr(err, 201)

	client := elasticloadbalancingv2.NewFromConfig(cfg)

	var arns []string
	if args := flag.Args(); isatty.IsTerminal(os.Stdin.Fd()) && len(args) > 0 {
		if isVerbose {
			fmt.Fprintln(os.Stderr, "Environments:", len(args))
		}
		arns, err = filterArns(ctx, client, args)
		ut.IsErr(err, 202, "filterArns()")
	} else {
		arns, err = readArns(os.Stdin)
		ut.IsErr(err, 203, "readArns()")
	}

	if isVerbose {
		fmt.Fprintln(os.Stderr, "Load Balancers:", len(arns))
	}

	if printUnhealthy(ctx, client, os.Stdout, arns) > 0 {
		os.Exit(1)
		// spew.Dump(arns) //
	}

	if isVerbose {
		fmt.Fprintln(os.Stderr, "Time spent:", strconv.FormatFloat(time.Since(start).Seconds(), 'f', 1, 64), "seconds")
	}
}
