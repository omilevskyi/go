package main

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/davecgh/go-spew/spew"

	"github.com/mattn/go-isatty"
	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	timeoutSec = 30

	maxDescribeTagsResources = 20 // DescribeTags accepts at most 20 resource ARNs per request

	eiSep = "-"
)

var eiTags = map[string]bool{
	"Environment":            true,
	"ei:environment":         true,
	"ei:provisioning-source": false,
}

type lbInfoT struct {
	env string
	idx int
}

func arnsByTags(ctx context.Context, c *elasticloadbalancingv2.Client, tags map[string]bool, arns map[string]lbInfoT) map[string]int {
	if len(arns) > 0 {
		out, err := c.DescribeTags(
			ctx, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: ut.Keys(arns)},
		)
		if err == nil {
			result := make(map[string]int, len(arns))
			for _, td := range out.TagDescriptions {
				if info, ok := arns[*td.ResourceArn]; ok {
					envPath := env2path(info.env)
					for _, tag := range td.Tags {
						if strict, ok := tags[*tag.Key]; ok && (strict && *tag.Value == info.env || !strict && strings.Contains(*tag.Value, envPath)) {
							result[*td.ResourceArn] = info.idx
							break
						}
					}
				}
			}
			if len(result) > 0 {
				return result
			}
		} else {
			ut.IsErr(err, -1)
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	ut.IsErr(err, 201)

	var arns []string
	if n := len(os.Args); isatty.IsTerminal(os.Stdin.Fd()) && n > 1 {
		client := elasticloadbalancingv2.NewFromConfig(cfg)
		lbs, err := describeLoadBalancers(ctx, client)
		ut.IsErr(err, 201)

		arns = make([]string, 0, (n-1)*len(lbs)/16) // estimatedLoadBalancerCount
		chunk := make(map[string]lbInfoT, maxDescribeTagsResources)

		for i := 1; i < n; i++ {
			for j := 0; j < len(lbs); j++ {
				if lbs[j].LoadBalancerArn != nil {
					if strings.HasPrefix(*lbs[j].LoadBalancerName, os.Args[i]+eiSep) || strings.HasSuffix(*lbs[j].LoadBalancerName, eiSep+os.Args[i]) {
						arns = append(arns, *lbs[j].LoadBalancerArn)
						lbs[j].LoadBalancerArn = nil
					} else {
						chunk[*lbs[j].LoadBalancerArn] = lbInfoT{env: os.Args[i], idx: j}
						if len(chunk) == maxDescribeTagsResources {
							for arn, idx := range arnsByTags(ctx, client, eiTags, chunk) {
								arns = append(arns, arn)
								lbs[idx].LoadBalancerArn = nil
							}
							clear(chunk)
						}
					}
				}
			}
		}
		arns = append(arns, ut.Keys(arnsByTags(ctx, client, eiTags, chunk))...)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if line := bytes.TrimSpace(scanner.Bytes()); len(line) > 0 { // line is valid until next scanner.Scan() call
				arns = append(arns, string(line))
			}
		}
		ut.IsErr(scanner.Err(), 201, "scanner.Err()")
	}
	// for i := 0; i < len(arns); i++ {
	// 	fmt.Println(arns[i])
	// }
	spew.Dump(arns)
	os.Exit(0)
}
