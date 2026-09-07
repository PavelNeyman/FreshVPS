#!/usr/bin/env bash
# Download & install sing-box binary for Linux (Debian/Armbian/etc.)
# shellcheck disable=SC2034

# install_singbox_binary [dest_dir=/usr/local/bin]
install_singbox_binary() {
  local dest="${1:-/usr/local/bin}"
  local arch tag url tmp dir
  arch="$(arch_go)"
  mkdir -p "${dest}"
  tag="$(curl -fsSL https://api.github.com/repos/SagerNet/sing-box/releases/latest | jq -r .tag_name)"
  [[ -n "${tag}" && "${tag}" != null ]] || die "Cannot resolve sing-box release"
  url="https://github.com/SagerNet/sing-box/releases/download/${tag}/sing-box-${tag#v}-linux-${arch}.tar.gz"
  tmp="$(mktemp -d)"
  info "Downloading sing-box ${tag} (${arch})"
  curl -fsSL "${url}" -o "${tmp}/sb.tgz"
  tar -xzf "${tmp}/sb.tgz" -C "${tmp}"
  dir="${tmp}/sing-box-${tag#v}-linux-${arch}"
  install -m 755 "${dir}/sing-box" "${dest}/sing-box"
  rm -rf "${tmp}"
  printf '%s\n' "${tag}" >"${FRESHVPS_STATE_DIR:-/var/lib/freshvps}/singbox-version" 2>/dev/null || true
}
