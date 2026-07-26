# Go language routines made for personal use

## Verification of the authenticity

```sh
export VERESION=0.0.2
cosign verify-blob \
  --certificate-identity https://github.com/omilevskyi/go/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle "unhealthy-tgs_${VERESION}.sha256.sigstore.json" \
  "https://github.com/omilevskyi/go/releases/download/v${VERESION}/unhealthy-tgs_${VERESION}.sha256"
```
