# kubectl-cnf

`kubectl-cnf` is a simple `kubectl` plugin that helps switch between current-contexts in multiple kubeconfigs within a terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

## Dependencies

### Required

* `fzf` - https://github.com/junegunn/fzf
* `bat` - https://github.com/sharkdp/bat

### Optional

* `xsel` (Linux x11 clipboard) - https://github.com/kfish/xsel
* `wl-copy` (Linux Wayland clipboard) - https://github.com/bugaevc/wl-clipboard

## Installation

1. Run `kubectl krew install cnf` or just place `kubectl-cnf` into the directory within your `PATH` (e.g. `~/.local/bin`)
2. Create the `~/.kube/configs` directory and place your kubeconfigs there (or create symlinks)

Alternatively, you can build a `go` [version](https://github.com/hedgieinsocks/kubectl-cnf/tree/main/go) of the plugin.

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE          | DEFAULT           | DETAILS                                                                                               |
|-------------------|-------------------|-------------------------------------------------------------------------------------------------------|
| `KCNF_DIR`        | `~/.kube/configs` | directory with kubeconfigs                                                                            |
| `KCNF_HEIGHT`     | `40%`             | selection menu height                                                                                 |
| `KCNF_NO_VERBOSE` |                   | do not print auxiliary messages                                                                       |
| `KCNF_NO_SHELL`   |                   | do not launch a subshell, instead print `export KUBECONFIG=PATH` to stdout                            |
| `KCNF_COPY_CLIP`  |                   | when `KCNF_NO_SHELL` is set, copy `export KUBECONFIG=PATH` to clipboard instead of printing to stdout |

## Usage

```
kubectl cnf helps switch between current-contexts in multiple kubeconfigs

Usage:
  kubectl cnf [-h] [<string>]

Flags:
  -h, --help    show this message

Dependencies:
  fzf - https://github.com/junegunn/fzf
  bat - https://github.com/sharkdp/bat

  xsel (Linux x11 clipboard) - https://github.com/kfish/xsel
  wl-copy (Linux Wayland clipboard) - https://github.com/bugaevc/wl-clipboard

Prerequisites:
  directory '/home/user/.kube/configs' populated with kubeconfig files
```

You can press `TAB` to preview the kubeconfig of the selected context.

## P.S.

If launching a subshell is an unacceptable overhead, you might want to try [zsh functions](https://github.com/hedgieinsocks/kubectl-cnf/blob/main/zsh.sh) instead.
