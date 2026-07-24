package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

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
