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
