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

func arnsByTags(ctx context.Context, c *elasticloadbalancingv2.Client, arns []string, tags map[string]bool, envName string) []string {
	if len(arns) > 0 {
		out, err := c.DescribeTags(
			ctx, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: arns},
		)
		if err == nil {
			var result []string
			envPath := env2path(envName)
			for _, td := range out.TagDescriptions {
				for _, tag := range td.Tags {
					if strict, ok := tags[*tag.Key]; ok {
						if strict && *tag.Value == envName || !strict && strings.Contains(*tag.Value, envPath) {
							result = append(result, *td.ResourceArn)
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
		chnkArns := make([]string, 0, maxDescribeTagsResources)

		for j := 0; j < len(lbs); j++ {
			for i := 1; i < n; i++ {
				if strings.HasPrefix(*lbs[j].LoadBalancerName, os.Args[i]+eiSep) || strings.HasSuffix(*lbs[j].LoadBalancerName, eiSep+os.Args[i]) {
					arns = append(arns, *lbs[j].LoadBalancerArn)
				} else {
					chnkArns = append(chnkArns, *lbs[j].LoadBalancerArn)
					if len(chnkArns) == cap(chnkArns) {
						arns = append(arns, arnsByTags(context.Background(), client, chnkArns, eiTags, os.Args[i])...)
						chnkArns = chnkArns[:0]
					}
				}
				arns = append(arns, arnsByTags(context.Background(), client, chnkArns, eiTags, os.Args[i])...)
				chnkArns = chnkArns[:0]
			}
		}
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
