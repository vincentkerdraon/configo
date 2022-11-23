# How to publish the doc on https://pkg.go.dev/

```bash
# git tag

# This will take some time to show up on https://pkg.go.dev/
# It will also show the go sub-modules
GOPROXY=https://proxy.golang.org go list -m github.com/vincentkerdraon/configo@v0.2.0
```
