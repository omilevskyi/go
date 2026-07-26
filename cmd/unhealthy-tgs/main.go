package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/mattn/go-isatty"
	ut "github.com/omilevskyi/go/pkg/utils"
	// "github.com/davecgh/go-spew/spew"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	ut.IsErr(err, 201)

	client := elasticloadbalancingv2.NewFromConfig(cfg)

	var arns []string
	if isatty.IsTerminal(os.Stdin.Fd()) && len(os.Args) > 1 {
		arns, err = filterArns(ctx, client, os.Args[1:])
		ut.IsErr(err, 201, "filterArns()")
	} else {
		arns, err = readArns(os.Stdin)
		ut.IsErr(err, 201, "readArns()")
	}

	if printUnhealthy(ctx, client, os.Stdout, arns) > 0 {
		os.Exit(1)
		// spew.Dump(arns)
	}
}
