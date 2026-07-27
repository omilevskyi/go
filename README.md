# Go language routines made for personal use

## Verification of the authenticity

```sh
export VERSION=0.0.2
cosign verify-blob \
  --certificate-identity https://github.com/omilevskyi/go/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle "unhealthy-tgs_${VERSION}.sha256.sigstore.json" \
  "https://github.com/omilevskyi/go/releases/download/v${VERSION}/unhealthy-tgs_${VERSION}.sha256"
```
