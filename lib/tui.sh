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

# tui_menu "title" tag1 "desc1" tag2 "desc2" ... → prints selected tag
tui_menu() {
  local title="$1" backend; shift
  backend="$(tui_detect)"
  if [[ "${NONINTERACTIVE:-0}" -eq 1 ]]; then
    printf '%s\n' "$1"
    return 0
  fi
  case "${backend}" in
    whiptail)
      whiptail --title "${title}" --menu "Select" 20 72 10 "$@" 3>&1 1>&2 2>&3
      ;;
    dialog)
      dialog --title "${title}" --menu "Select" 20 72 10 "$@" 3>&1 1>&2 2>&3
      ;;
    *)
      local i=1 tags=() descs=() t d n
      while [[ $# -ge 2 ]]; do tags+=("$1"); descs+=("$2"); shift 2; done
      echo "${title}"
      for i in "${!tags[@]}"; do
        printf '  %d) %s — %s\n' "$((i + 1))" "${tags[$i]}" "${descs[$i]}"
      done
      read -r -p "> " n || true
      n="${n:-1}"
      if [[ "${n}" =~ ^[0-9]+$ ]] && [[ "${n}" -ge 1 && "${n}" -le ${#tags[@]} ]]; then
        printf '%s\n' "${tags[$((n - 1))]}"
      else
        printf '%s\n' "${tags[0]}"
      fi
      ;;
  esac
}

# tui_checklist "title" tag1 "desc1" on|off|0|1  tag2 ... → space-separated selected tags
tui_checklist() {
  local title="$1" backend; shift
  backend="$(tui_detect)"
  if [[ "${NONINTERACTIVE:-0}" -eq 1 ]]; then
    # return all marked on/1
    local out=() t d s
    while [[ $# -ge 3 ]]; do
      t="$1"; d="$2"; s="$3"; shift 3
      [[ "${s}" == "on" || "${s}" == "1" ]] && out+=("${t}")
    done
    printf '%s\n' "${out[*]}"
    return 0
  fi
  case "${backend}" in
    whiptail)
      # rebuild args as tag desc status
      local args=() t d s
      while [[ $# -ge 3 ]]; do
        t="$1"; d="$2"; s="$3"; shift 3
        [[ "${s}" == "1" ]] && s=on
        [[ "${s}" == "0" ]] && s=off
        args+=("${t}" "${d}" "${s}")
      done
      # shellcheck disable=SC2068
      whiptail --title "${title}" --checklist "Space=toggle  Enter=OK" 22 78 12 "${args[@]}" 3>&1 1>&2 2>&3 | tr -d '"'
      ;;
    dialog)
      local args=() t d s
      while [[ $# -ge 3 ]]; do
        t="$1"; d="$2"; s="$3"; shift 3
        [[ "${s}" == "1" ]] && s=on
        [[ "${s}" == "0" ]] && s=off
        args+=("${t}" "${d}" "${s}")
      done
      dialog --title "${title}" --checklist "Select" 22 78 12 "${args[@]}" 3>&1 1>&2 2>&3 | tr -d '"'
      ;;
    *)
      local t d s a out=()
      echo "${title} (y/n for each)"
      while [[ $# -ge 3 ]]; do
        t="$1"; d="$2"; s="$3"; shift 3
        def=n; [[ "${s}" == "on" || "${s}" == "1" ]] && def=y
        read -r -p "  ${t} — ${d} [${def}]: " a || true
        a="${a:-${def}}"
        [[ "${a}" =~ ^[Yy] ]] && out+=("${t}")
      done
      printf '%s\n' "${out[*]}"
      ;;
  esac
}
