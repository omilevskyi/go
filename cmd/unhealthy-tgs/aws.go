package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"golang.org/x/sync/errgroup"

	"github.com/omilevskyi/go/pkg/aws"
	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	maxDescribeTagsResources = 20 // DescribeTags accepts at most 20 resource ARNs per request
	maxDescribeTargetGroups  = 3
	maxDescribeTargetHealth  = 6

	eiSep = "-"
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

// tagsMatchEnv determines whether a resource belongs to the specified
// environment based on its tags.
//
// Environment tags are evaluated according to the matching rules defined
// in eiTags. Some tags require an exact environment match, while others
// are matched against the environment path format produced by env2path().
func tagsMatchEnv(eiTags map[string]bool, tags []types.Tag, env string) bool {
	envPath := env2path(env)
	for _, tag := range tags {
		if strict, ok := eiTags[*tag.Key]; ok &&
			(strict && *tag.Value == env || !strict && strings.Contains(*tag.Value, envPath)) {
			return true
		}
	}
	return false
}

// findMatchingArns returns ARNs of load balancers whose tags match
// the environment information associated with each load balancer.
func findMatchingArns(ctx context.Context, c *elasticloadbalancingv2.Client,
	eiTags map[string]bool, arns map[string]lbInfoT, cachedTags *map[string][]types.Tag,
) map[string]int {
	if len(arns) > 0 {
		ctx1, cancel := context.WithTimeout(ctx, timeoutSec*time.Second) // ctx1 is used once
		out, err := c.DescribeTags(
			ctx1, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: ut.Keys(arns)},
		)
		cancel()
		if err == nil {
			result := make(map[string]int, len(arns))
			for _, td := range out.TagDescriptions {
				if v, ok := arns[*td.ResourceArn]; ok && tagsMatchEnv(eiTags, td.Tags, v.env) {
					result[*td.ResourceArn] = v.idx
				} else if cachedTags != nil {
					(*cachedTags)[*td.ResourceArn] = td.Tags
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

func processTargetHealth(ctx context.Context, c *elasticloadbalancingv2.Client, lb LbT, tgARN string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec*time.Second)
	ths, err := aws.DescribeTargetHealth(ctx, c, tgARN)
	cancel()
	if err != nil {
		return err
	}
	for _, th := range ths {
		if th.TargetHealth.State == types.TargetHealthStateEnumHealthy {
			return nil
		}
	}
	if len(ths) > 0 {
		lb.mu.Lock()
		*lb.TgARNs = append(*lb.TgARNs, tgARN)
		lb.mu.Unlock()
	}
	return nil
}

func processTargetGroups(ctx context.Context, c *elasticloadbalancingv2.Client, lb LbT) error {
	ctx1, cancel := context.WithTimeout(ctx, timeoutSec*time.Second) // ctx1 is used once
	tgs, err := aws.DescribeTargetGroups(ctx1, c, lb.ARN)
	cancel()
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxDescribeTargetHealth)
	for _, tg := range tgs {
		g.Go(func() error {
			return processTargetHealth(ctx, c, lb, *tg.TargetGroupArn)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func selectUnhealthy(ctx context.Context, profiles []ProfileT) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxDescribeTargetGroups)
	for _, p := range profiles {
		for _, env := range *p.envs {
			for _, lb := range *env.LBs {
				g.Go(func() error {
					return processTargetGroups(ctx, p.ELBv2Client, lb)
				})
			}
		}
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// groupLoadBalancers discovers load balancers associated with the given
// environments and returns their ARNs. A load balancer is considered a match
// if its name contains the environment identifier or if its tags satisfy
// the configured environment matching rules. When no environments are
// specified, all load balancer ARNs are returned.
func groupLoadBalancers(ctx context.Context, profiles []ProfileT) error {
	g, ctx := errgroup.WithContext(ctx)
	for i := range profiles {
		g.Go(func() error {
			return grpLBsInProfile(ctx, &profiles[i])
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func grpLBsInProfile(ctx context.Context, p *ProfileT) error {
	verbf(vInfo, "%s: start", p.Name)
	ctx1, cancel := context.WithTimeout(ctx, timeoutSec*time.Second) // ctx1 is used once
	lbs, err := aws.DescribeLoadBalancers(ctx1, p.ELBv2Client)
	cancel()
	if err != nil {
		return err
	}
	m, n := len(*p.envs), len(lbs)
	verbf(vNotice, "%s: load balancers: %d", p.Name, n)
	if m > 0 {
		if n > 0 {
			cacheTags := make(map[string][]types.Tag, n)
			arns := make([]string, n/10) // empirical value
			chunk := make(map[string]lbInfoT, maxDescribeTagsResources)
			for i := range m { // per environment
				arns = arns[:0]
				verbf(vInfo, "%s: %s: len(arns)=%d, cap(arns)=%d, cache=%d", p.Name, (*p.envs)[i].Name, len(arns), cap(arns), len(cacheTags))
				for j := range n { // per load balancer
					if lbs[j].LoadBalancerArn != nil {

						if strings.HasPrefix(*lbs[j].LoadBalancerName, (*p.envs)[i].Name+eiSep) ||
							strings.HasSuffix(*lbs[j].LoadBalancerName, eiSep+(*p.envs)[i].Name) {
							verbf(vDebug, "%s: name: %s", p.Name, *lbs[j].LoadBalancerName)
							delete(cacheTags, *lbs[j].LoadBalancerArn)
							arns = append(arns, *lbs[j].LoadBalancerArn)
							lbs[j].LoadBalancerArn = nil
							continue
						}

						if tags, ok := cacheTags[*lbs[j].LoadBalancerArn]; ok &&
							tagsMatchEnv(eiTags, tags, (*p.envs)[i].Name) {
							verbf(vDebug, "%s: %s: hit: %s", p.Name, (*p.envs)[i].Name, *lbs[j].LoadBalancerName)
							delete(cacheTags, *lbs[j].LoadBalancerArn)
							arns = append(arns, *lbs[j].LoadBalancerArn)
							lbs[j].LoadBalancerArn = nil
							continue
						}

						chunk[*lbs[j].LoadBalancerArn] = lbInfoT{env: (*p.envs)[i].Name, idx: j}
						if len(chunk) == maxDescribeTagsResources {
							for arn, l := range findMatchingArns(ctx, p.ELBv2Client, eiTags, chunk, &cacheTags) {
								verbf(vDebug, "%s: %s: tags: %s", p.Name, (*p.envs)[i].Name, *lbs[l].LoadBalancerArn)
								delete(cacheTags, arn)
								arns = append(arns, arn)
								lbs[l].LoadBalancerArn = nil
							}
							clear(chunk)
						}
					}
				}
				arns = append(arns, ut.Keys(findMatchingArns(ctx, p.ELBv2Client, eiTags, chunk, &cacheTags))...)
				(*p.envs)[i].LBs = new(make([]LbT, len(arns)))
				for j := range arns {
					(*(*p.envs)[i].LBs)[j] = LbT{arns[j], new(make([]string, 0)), new(sync.Mutex)}
				}
				verbf(vNotice, "%s: %s: load balancers matched: %d", p.Name, (*p.envs)[i].Name, len(*(*p.envs)[i].LBs))
			}
		}
	} else {
		p.envs = &[]EnvT{{"", new(make([]LbT, n))}}
		for j := range n {
			(*(*p.envs)[0].LBs)[j] = LbT{*lbs[j].LoadBalancerArn, new(make([]string, 0)), new(sync.Mutex)}
		}
	}
	verbf(vInfo, "%s: finish", p.Name)
	return nil
}
