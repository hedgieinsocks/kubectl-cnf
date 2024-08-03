# kubectl-cnf

`kubectl-cnf` is a simple `kubectl` plugin that helps switch between current-contexts in multiple kubeconfigs within a terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

## Dependencies

* `fzf` - https://github.com/junegunn/fzf
* `bat` - https://github.com/sharkdp/bat

## Installation

1. Run `kubectl krew install cnf` or just place `kubectl-cnf` into the directory within your `PATH` (e.g. `~/.local/bin`)
2. Create the `~/.kube/configs` directory and place your kubeconfigs there (or create symlinks)

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE        | DEFAULT           | DETAILS                                                  |
|-----------------|-------------------|----------------------------------------------------------|
| `KCNF_DIR`      | `~/.kube/configs` | directory with your kubeconfigs                          |
| `KCNF_VERBOSE`  | `true`            | print a notificaiton when entering or exiting a subshell |
| `KCNF_SUBSHELL` | `true`            | export `KUBECONFIG` and launch a subshell                |
| `KCNF_HEIGHT`   | `40%`             | fzf height                                               |

## Usage

```
kubectl cnf helps switch between current-contexts in multiple kubeconfigs

Usage:
  kubectl cnf [-h] [<string>]

Flags:
  -h, --help    show this message
```

You can press `TAB` to preview the kubeconfig of the selected context.

## P.S.

If launching a subshell is an unacceptable overhead, you might want to try [zsh functions](https://github.com/hedgieinsocks/kubectl-cnf/blob/main/zsh.sh) instead.
