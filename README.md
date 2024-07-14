# kubectl-cnf

`kubectl-cnf` is a simple `kubectl` plugin that allows one to switch between multiple k8s configs within the terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

## Dependencies

* `fzf` - https://github.com/junegunn/fzf
* `bat` - https://github.com/sharkdp/bat

## Installation

1. Run `kubectl krew install cnf` or just place `kubectl-cnf` into the directory within your `PATH` (e.g. `~/.local/bin`)
2. Create the `~/.kube/configs` directory and place your kubeconfigs there (or create symlinks)

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE        | DEFAULT           | DETAILS                                                |
|-----------------|-------------------|--------------------------------------------------------|
| `KCNF_DIR`      | `~/.kube/configs` | directory with your kubeconfigs                        |
| `KCNF_VERBOSE`  | `true`            | print a notificaiton when entering or exiting subshell |
| `KCNF_SUBSHELL` | `true`            | export `KUBECONFIG` and launch subshell                |
| `KCNF_HEIGHT`   | `40%`             | fzf height                                             |

## P.S.

You can opt in for the shell funciton from `kcnf.sh` to avoid the plugin overhead.
