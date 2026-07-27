package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	timeoutSec               = 30
	maxDescribeTagsResources = 20 // DescribeTags accepts at most 20 resource ARNs per request
	eiSep                    = "-"
)

type lbInfoT struct {
	env string
	idx int
}

var eiTags = map[string]bool{
	"Environment":            true,
	"ei:environment":         true,
	"ProvisioningSource":     false,
	"ei:provisioning-source": false,
}

// describeLoadBalancers hides ELBv2 pagination and returns all load balancers
// as a single slice.
func describeLoadBalancers(ctx context.Context, c *elasticloadbalancingv2.Client) ([]types.LoadBalancer, error) {
	var lbs []types.LoadBalancer

	p := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(
		c, &elasticloadbalancingv2.DescribeLoadBalancersInput{},
	)

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		lbs = append(lbs, page.LoadBalancers...)
	}

	if len(lbs) > 0 {
		return lbs, nil
	}
	return nil, nil
}

// describeTargetGroups hides ELBv2 pagination and returns all target groups
// as a single slice.
func describeTargetGroups(ctx context.Context, c *elasticloadbalancingv2.Client, arn string) ([]types.TargetGroup, error) {
	var tgs []types.TargetGroup

	p := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(
		c, &elasticloadbalancingv2.DescribeTargetGroupsInput{LoadBalancerArn: &arn},
	)

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		tgs = append(tgs, page.TargetGroups...)
	}

	if len(tgs) > 0 {
		return tgs, nil
	}
	return nil, nil
}

// describeTargetHealth retrieves the registered targets of a target group and
// their health information.
func describeTargetHealth(ctx context.Context, c *elasticloadbalancingv2.Client, arn string) ([]types.TargetHealthDescription, error) {
	out, err := c.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: &arn,
	})
	if err != nil {
		return nil, err
	}

	if len(out.TargetHealthDescriptions) > 0 {
		return out.TargetHealthDescriptions, nil
	}
	return nil, nil
}

// arnsByTags returns ARNs of load balancers whose tags match the environment
// information associated with each load balancer.
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

// filterArns discovers load balancers associated with the given environments
// and returns their ARNs. A load balancer is considered a match if its name
// contains the environment identifier or if its tags satisfy the configured
// environment matching rules. When no environments are specified, all load
// balancer ARNs are returned.
func filterArns(ctx context.Context, c *elasticloadbalancingv2.Client, envs []string) ([]string, error) {
	lbs, err := describeLoadBalancers(ctx, c)
	if err != nil {
		return nil, err
	}

	var arns []string
	if n := len(lbs); n > 0 {
		if m := len(envs); m > 0 {
			chunk := make(map[string]lbInfoT, maxDescribeTagsResources)
			arns = make([]string, 0, m*n/16)
			for i := 0; i < m; i++ {
				for j := 0; j < n; j++ {
					if lbs[j].LoadBalancerArn != nil {
						if strings.HasPrefix(*lbs[j].LoadBalancerName, envs[i]+eiSep) || strings.HasSuffix(*lbs[j].LoadBalancerName, eiSep+envs[i]) {
							arns = append(arns, *lbs[j].LoadBalancerArn)
							lbs[j].LoadBalancerArn = nil
						} else {
							chunk[*lbs[j].LoadBalancerArn] = lbInfoT{env: envs[i], idx: j}
							if len(chunk) == maxDescribeTagsResources {
								for arn, idx := range arnsByTags(ctx, c, eiTags, chunk) {
									arns = append(arns, arn)
									lbs[idx].LoadBalancerArn = nil
								}
								clear(chunk)
							}
						}
					}
				}
			}
			arns = append(arns, ut.Keys(arnsByTags(ctx, c, eiTags, chunk))...)
		} else {
			arns = make([]string, 0, n)
			for j := 0; j < n; j++ {
				arns = append(arns, *lbs[j].LoadBalancerArn)
			}
		}
	}

	if len(arns) > 0 {
		return ut.Arrange(arns), nil
	}
	return nil, nil
}

// printUnhealthy checks all target groups associated with the specified load
// balancers and reports target groups that contain targets but none in the
// Healthy state. It returns the total number of such target groups.
func printUnhealthy(ctx context.Context, c *elasticloadbalancingv2.Client, w io.Writer, arns []string) int {
	count := 0
	for i := 0; i < len(arns); i++ {
		tgs, err := describeTargetGroups(ctx, c, arns[i])
		if err == nil {
			printed := false
			for _, tg := range tgs {
				ths, err := describeTargetHealth(ctx, c, *tg.TargetGroupArn)
				if err != nil {
					ut.IsErr(err, -1)
					continue
				}
				if len(ths) > 0 {
					healthy := false
					for _, th := range ths {
						if th.TargetHealth.State == types.TargetHealthStateEnumHealthy {
							healthy = true
							break
						}
					}
					if !healthy {
						if !printed {
							fmt.Fprintln(w, arns[i])
							printed = true
						}
						fmt.Fprintln(w, " ", *tg.TargetGroupArn, "unhealthy")
						count++
					}
				}
			}
			if printed {
				fmt.Fprintln(w)
			}
		} else {
			ut.IsErr(err, -1)
		}
	}
	return count
}
