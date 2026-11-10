module github.com/omilevskyi/go/unhealthy-tgs

go 1.27

require (
	github.com/aws/aws-sdk-go-v2 v1.47.7
	github.com/aws/aws-sdk-go-v2/config v1.33.13
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.63.8
	github.com/go-jose/go-jose/v4 v4.1.9
	github.com/go-openapi/testify/v2 v2.8.0
	github.com/mattn/go-isatty v0.0.24
	github.com/omilevskyi/go v0.1.42
	golang.org/x/sync v0.24.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.20.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.39.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.44.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.6 // indirect
	github.com/aws/smithy-go v1.28.8 // indirect
	golang.org/x/sys v0.50.0 // indirect
)

replace github.com/omilevskyi/go => github.com/omilevskyi/go v0.1.36
