module github.com/omilevskyi/go/unhealthy-tgs

go 1.26

require (
	github.com/aws/aws-sdk-go-v2 v1.45.1
	github.com/aws/aws-sdk-go-v2/config v1.33.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.60.1
	github.com/go-jose/go-jose/v4 v4.1.4
	github.com/go-openapi/testify/v2 v2.7.0
	github.com/mattn/go-isatty v0.0.24
	github.com/omilevskyi/go v0.1.24
	golang.org/x/sync v0.22.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.40.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.47.1 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/omilevskyi/go => github.com/omilevskyi/go v0.1.19
