#!/usr/bin/env bash
# Portable TUI helpers — no required packages.
# Backends: whiptail | dialog | osascript (macOS) | plain read
# shellcheck disable=SC2034

tui_detect() {
  if [[ -n "${FRESHVPS_TUI:-}" ]]; then
    echo "${FRESHVPS_TUI}"
    return
  fi
  case "$(uname -s)" in
    Darwin)
      if command -v osascript >/dev/null 2>&1; then echo osascript; return; fi
      ;;
  esac
  if command -v whiptail >/dev/null 2>&1; then echo whiptail; return; fi
  if command -v dialog >/dev/null 2>&1; then echo dialog; return; fi
  echo plain
}

# tui_yesno "question" default_y|n  → 0 yes / 1 no
tui_yesno() {
  local q="$1" def="${2:-y}" backend
  backend="$(tui_detect)"
  if [[ "${NONINTERACTIVE:-0}" -eq 1 ]]; then
    [[ "${def}" == "y" ]] && return 0 || return 1
  fi
  case "${backend}" in
    whiptail)
      if [[ "${def}" == "y" ]]; then whiptail --yesno "${q}" 10 60; else whiptail --defaultno --yesno "${q}" 10 60; fi
      ;;
    dialog)
      if [[ "${def}" == "y" ]]; then dialog --yesno "${q}" 10 60; else dialog --defaultno --yesno "${q}" 10 60; fi
      ;;
    osascript)
      local ans
      ans="$(osascript -e "display dialog \"${q}\" buttons {\"No\",\"Yes\"} default button \"$([ "${def}" = y ] && echo Yes || echo No)\"" -e 'button returned of result' 2>/dev/null || true)"
      [[ "${ans}" == "Yes" ]]
      ;;
    *)
      local a
      read -r -p "${q} [$([ "${def}" = y ] && echo Y/n || echo y/N)] " a || true
      a="${a:-${def}}"
      [[ "${a}" =~ ^[Yy] ]]
      ;;
  esac
}

# tui_input "question" "default" → prints value
tui_input() {
  local q="$1" def="${2:-}" backend
  backend="$(tui_detect)"
  if [[ "${NONINTERACTIVE:-0}" -eq 1 ]]; then
    printf '%s\n' "${def}"
    return
  fi
  case "${backend}" in
    whiptail)
      whiptail --inputbox "${q}" 10 60 "${def}" 3>&1 1>&2 2>&3 || printf '%s\n' "${def}"
      ;;
    dialog)
      dialog --inputbox "${q}" 10 60 "${def}" 3>&1 1>&2 2>&3 || printf '%s\n' "${def}"
      ;;
    osascript)
      osascript -e "text returned of (display dialog \"${q}\" default answer \"${def}\")" 2>/dev/null || printf '%s\n' "${def}"
      ;;
    *)
      local a
      read -r -p "${q} [${def}]: " a || true
      printf '%s\n' "${a:-${def}}"
      ;;
  esac
}

# tui_msg "text"
tui_msg() {
  local t="$1" backend
  backend="$(tui_detect)"
  case "${backend}" in
    whiptail) whiptail --msgbox "${t}" 12 70 ;;
    dialog) dialog --msgbox "${t}" 12 70 ;;
    osascript) osascript -e "display dialog \"${t}\" buttons {\"OK\"} default button 1" >/dev/null 2>&1 || echo "${t}" ;;
    *) echo "${t}" ;;
  esac
}
