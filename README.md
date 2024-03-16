# kubectl-cnf

`kubectl-cnf` is a `kubectl` plugin that allows one to switch between multiple k8s configs within the terminal tab scope.

If you are working with many clusters that come with their own kubeconfigs, you can use this tool to allow yourself multitask by keeping a dedicated terminal tab for each cluster.

## Dependencies

* `kubectl` - https://kubernetes.io/docs/tasks/tools/install-kubectl
* `fzf` - https://github.com/junegunn/fzf

## Installation

1. Install the dependencies from above
2. Place `kubectl-cnf` into the directory that is a part of your `PATH` (e.g. `~/.krew/bin`)
3. Create the `~/.kube/configs` directory and place your kubeconfigs there (or create symlinks)

## Customization

You can export the following variables to tweak the plugin's behaviour.

| VARIABLE       | DEFAULT           |
|----------------|-------------------|
| `KCNF_DIR`     | `~/.kube/configs` |
| `KCNF_VERBOSE` | `true`            |

## P.S.

To avoid using subshells, you can opt in for a shell funciton from `kcnf.sh`
