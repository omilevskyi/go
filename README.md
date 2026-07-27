# Go language routines made for personal use

[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://raw.githubusercontent.com/omilevskyi/go/refs/heads/main/LICENSE)
[![Powered By: GoReleaser](https://img.shields.io/badge/Powered%20by-GoReleaser-blue.svg)](https://goreleaser.com/)
[![Powered By: Cosign](https://img.shields.io/badge/Powered%20by-Cosign-blue.svg)](https://docs.sigstore.dev/quickstart/quickstart-cosign/)

## Verification of the authenticity

```sh
export VERSION=0.1.2
cosign verify-blob \
  --certificate-identity https://github.com/omilevskyi/go/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle "unhealthy-tgs_${VERSION}.sha256.sigstore.json" \
  "https://github.com/omilevskyi/go/releases/download/v${VERSION}/unhealthy-tgs_${VERSION}.sha256"
```
