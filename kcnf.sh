#!/usr/bin/env bash

kcnf() {
  local config

  hash fzf bat || return 1
  mkdir -p "${KCNF_DIR:-$HOME/.kube/configs}"

  config=$(find "${KCNF_DIR:-$HOME/.kube/configs}" \( -type f -o -type l \) -exec awk '/^current-context:/ {print FILENAME, $2}' {} + \
    | sort -k2 \
    | fzf \
      --cycle \
      --height="${KCNF_HEIGHT:-40%}" \
      --layout=reverse \
      --with-nth=2 \
      --query="${*:-}" \
      --bind="tab:toggle-preview" \
      --preview="bat --style=plain --color=always --language=yaml {1}" \
      --preview-window="hidden,wrap,75%")

  if [[ -n "${config}" ]]; then
    export KUBECONFIG="${config%% *}"
    export KUBECONTEXT="${config##* }"
  fi

  return 0
}
