#!/usr/bin/env bash

kcnf() {
  local config

  hash fzf || return 1

  config=$(find "${KCNF_DIR:-$HOME/.kube/configs}" \( -type f -o -type l \) -exec awk '/^current-context:/ {print FILENAME, $2}' {} + \
    | sort -k2 \
    | fzf --cycle --layout=reverse --query "${*:-}" --with-nth=2)

  [[ -n "${config}" ]] || return

  export KUBECONFIG="${config%% *}"
  export KUBECONTEXT="${config##* }"
}
