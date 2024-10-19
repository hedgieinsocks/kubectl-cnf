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
  -q, --quiet              do not print auxiliary messages
  -p, --print              print 'export KUBECONFIG=PATH' to stdout instead of launching a subshell
  -c, --clipboard          copy 'export KUBECONFIG=PATH' to clipboard instead of launching a subshell
  -l, --link               symlink selected kubeconfig to '~/.kube/config' instead of launching a subshell
  -h, --help               help for kubectl cnf
  -v, --version            version for kubectl cnf
  ```
