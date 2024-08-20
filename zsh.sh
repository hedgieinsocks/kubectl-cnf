#!/usr/bin/env zsh

_set_kubectx_title() {
  if [[ -n "${KUBECONFIG}" ]]; then
    export DISABLE_AUTO_TITLE="true"
    # oh-my-zsh https://github.com/ohmyzsh/ohmyzsh/blob/master/lib/termsupport.zsh#L9
    title "☸ ${KUBECONTEXT}"
  fi
}
precmd_functions+=(_set_kubectx_title)

kcnf() {
  local selected_kubeconfig

  selected_kubeconfig=$(find ~/.kube/configs \( -type f -o -type l \) -exec awk '/^current-context:/ {print FILENAME, $2}' {} + \
    | sort --key=2 \
    | fzf \
      --height="40%" \
      --layout=reverse \
      --with-nth=2 \
      --query="${*:-}" \
      --bind="tab:toggle-preview" \
      --preview="bat --style=header --color=always --language=yaml {1}" \
      --preview-window="hidden,wrap,75%")

  if [[ -n "${selected_kubeconfig}" ]]; then
    export KUBECONFIG="${selected_kubeconfig%% *}"
    export KUBECONTEXT="${selected_kubeconfig##* }"
  fi

  [[ -n "${ZLE_STATE}" ]] && zle reset-prompt && _set_kubectx_title
  return 0
}
zle -N kcnf
bindkey "^k" kcnf
