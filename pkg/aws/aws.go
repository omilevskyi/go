package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// DescribeLoadBalancers hides ELBv2 pagination and returns all load balancers
// as a single slice.
func DescribeLoadBalancers(ctx context.Context, c *elasticloadbalancingv2.Client) ([]types.LoadBalancer, error) {
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

// DescribeTargetGroups hides ELBv2 pagination and returns all target groups
// as a single slice.
func DescribeTargetGroups(ctx context.Context, c *elasticloadbalancingv2.Client, arn string) ([]types.TargetGroup, error) {
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

// DescribeTargetHealth retrieves the registered targets of a target group and
// their health information.
func DescribeTargetHealth(ctx context.Context, c *elasticloadbalancingv2.Client, arn string) ([]types.TargetHealthDescription, error) {
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
