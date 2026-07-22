package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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
	timeoutSec = 60

	maxDescribeTagsResources = 20 // DescribeTags accepts at most 20 resource ARNs per request

	eiEnvValue = "devo1"
)

var eiEnvKeys = []string{
	"ei:environment",
	"Environment",
}

func arnsWithTag(ctx context.Context, c *elasticloadbalancingv2.Client, arns []string, envName string, keys ...string) []string {
	if len(arns) > 0 {
		out, err := c.DescribeTags(
			ctx, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: arns},
		)
		if err == nil {
			var result []string
			for _, td := range out.TagDescriptions {
				spew.Dump(td)
			outer:
				for _, tag := range td.Tags {
					for _, key := range keys {
						if key == *tag.Key && envName == *tag.Value || *tag.Key == "ei:provisioning-source" && strings.Contains(*tag.Value, env2path(envName)) {
							fmt.Println(" ", *tag.Key, *tag.Value)
							result = append(result, *td.ResourceArn)
							break outer
						}
					}
				}
				fmt.Println()
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

	// fmt.Printf("#%v\n", cfg)
	// fmt.Println(string(ut.Must(json.MarshalIndent(cfg, "", "  "))))
	// spew.Dump(cfg)

	var arns []string

	if isatty.IsTerminal(os.Stdin.Fd()) {
		client := elasticloadbalancingv2.NewFromConfig(cfg)
		out, err := client.DescribeLoadBalancers(
			ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{},
		)
		ut.IsErr(err, 201)

		arns = make([]string, 0, len(out.LoadBalancers))
		chunkArns := make([]string, 0, maxDescribeTagsResources)
		for i := 0; i < len(out.LoadBalancers); i++ {
			chunkArns = append(chunkArns, *out.LoadBalancers[i].LoadBalancerArn)
			if len(chunkArns) == cap(chunkArns) {
				// spew.Dump(chunkArns)
				// fmt.Println()
				arns = append(arns, arnsWithTag(ctx, client, chunkArns, eiEnvValue, eiEnvKeys...)...)
				chunkArns = chunkArns[:0]
			}
		}
		arns = append(arns, arnsWithTag(ctx, client, chunkArns, eiEnvValue, eiEnvKeys...)...)
		// spew.Dump(chunkArns)
		// fmt.Println()
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if line := bytes.TrimSpace(scanner.Bytes()); len(line) > 0 { // line is valid until next scanner.Scan() call
				arns = append(arns, string(line))
			}
		}
		ut.IsErr(scanner.Err(), 201, "scanner.Err()")
	}
	spew.Dump(arns)
}
