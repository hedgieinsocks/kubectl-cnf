# kubectl-cnf

`kubectl-cnf` is a `kubectl` plugin that allows one to switch between multiple k8s configs within the terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

## Dependencies

* `fzf` - https://github.com/junegunn/fzf

## Installation

1. Install `fzf`
2. Place `kubectl-cnf` into the directory that is a part of your `PATH` (e.g. `~/.krew/bin`)
3. Create the `~/.kube/configs` directory and place your kubeconfigs there (or create symlinks)

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE        | DEFAULT           | DETAILS                                                |
|-----------------|-------------------|--------------------------------------------------------|
| `KCNF_DIR`      | `~/.kube/configs` | directory with your kubeconfigs                        |
| `KCNF_VERBOSE`  | `true`            | print a notificaiton when entering or exiting subshell |
| `KCNF_SUBSHELL` | `true`            | export `KUBECONFIG` and launch subshell                |

## P.S.

You can opt in for a shell funciton from `kcnf.sh` to avoid the plugin overhead.
