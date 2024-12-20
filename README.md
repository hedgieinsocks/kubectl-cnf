# kubectl-cnf

`kubectl-cnf` is a simple `kubectl` plugin that helps switch between current-contexts in multiple kubeconfigs within a terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

By default, the plugin launches a subshell for the chosen kubeconfig. But you can tweak the plugin to just print `export KUBECONFIG=PATH` command or copy it to the clipboard instead.

![kubectl-cnf demo GIF](img/demo-1.gif)

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

The krew version is a simple bash script, but you can grab a `go` [version](https://github.com/hedgieinsocks/kubectl-cnf/tree/main/go) of the plugin from the [releases](https://github.com/hedgieinsocks/kubectl-cnf/releases) page.

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE          | DEFAULT           | DETAILS                                                                         |
|-------------------|-------------------|---------------------------------------------------------------------------------|
| `KCNF_DIR`        | `~/.kube/configs` | directory with kubeconfigs                                                      |
| `KCNF_HEIGHT`     | `40%`             | selection menu height                                                           |
| `KCNF_NO_VERBOSE` | `0`               | do not print auxiliary messages                                                 |
| `KCNF_NO_SHELL`   | `0`               | print `export KUBECONFIG=PATH` to stdout instead of launching a subshell        |
| `KCNF_COPY_CLIP`  | `0`               | copy `export KUBECONFIG=PATH` to clipboard instead of launching a subshell      |
| `KCNF_SYMLINK`    | `0`               | symlink selected kubeconfig to `~/.kube/config` instead of launching a subshell |

## Usage

```
kubectl cnf helps switch between current-contexts in multiple kubeconfigs

Usage:
  kubectl cnf [-h] [<string>]

Flags:
  -h, --help        show this message
  -v, --version     show plugin version

Dependencies:
  fzf - https://github.com/junegunn/fzf
  bat - https://github.com/sharkdp/bat

Prerequisites:
  directory '/home/user/.kube/configs' populated with kubeconfig files
```

You can press `TAB` to preview the kubeconfig of the selected context.

## P.S.

If launching a subshell is an unacceptable overhead, you might want to try [zsh functions](https://github.com/hedgieinsocks/kubectl-cnf/blob/main/zsh.sh) instead.
