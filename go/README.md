This is just a `go` implementation of the plugin.

## Build

```sh
❯ go build -ldflags="-s -w"
```

## Usage

```
kubectl cnf helps switch between current-contexts in multiple kubeconfigs

Usage:
  kubectl cnf [flags]

Flags:
  -d, --directory string   directory with kubeconfigs (default "/home/user/.kube/configs")
  -H, --height string      selection menu height (default "40%")
  -V, --no-verbose         do not print auxiliary messages
  -S, --no-shell           do not launch a subshell, instead print 'export KUBECONFIG=PATH' to stdout
  -c, --clipboard          when --no-shell is provided, copy 'export KUBECONFIG=PATH' to clipboard instead of printing to stdout
  -h, --help               help for kubectl cnf
  -v, --version            version for kubectl-cnf
  ```
