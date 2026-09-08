module github.com/omilevskyi/go/unhealthy-tgs

go 1.27

require (
	github.com/aws/aws-sdk-go-v2 v1.46.0
	github.com/aws/aws-sdk-go-v2/config v1.33.3
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.62.0
	github.com/go-jose/go-jose/v4 v4.1.5
	github.com/go-openapi/testify/v2 v2.7.0
	github.com/mattn/go-isatty v0.0.24
	github.com/omilevskyi/go v0.1.28
	golang.org/x/sync v0.23.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.20.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.37.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.42.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.49.0 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

replace github.com/omilevskyi/go => github.com/omilevskyi/go v0.1.26
