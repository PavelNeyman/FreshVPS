#!/usr/bin/env bash
# Module: hardening
# shellcheck disable=SC2154

module_hardening_install() {
  info "Installing base hardening packages"
  pkg_install ufw fail2ban unattended-upgrades apt-listchanges needrestart \
    curl ca-certificates openssh-server

  if ! sysctl net.ipv4.tcp_congestion_control 2>/dev/null | grep -q bbr; then
    cat >/etc/sysctl.d/99-freshvps-bbr.conf <<'EOF'
net.core.default_qdisc=fq
net.ipv4.tcp_congestion_control=bbr
EOF
    sysctl --system >/dev/null 2>&1 || true
    info "Enabled BBR"
  else
    info "BBR already enabled"
  fi

  local sshd=/etc/ssh/sshd_config
  if [[ -f "${sshd}" ]]; then
    cp -a "${sshd}" "${sshd}.freshvps.bak.$(date +%s)" || true
    if grep -qE '^#?PasswordAuthentication' "${sshd}"; then
      sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' "${sshd}" || true
    else
      echo 'PasswordAuthentication no' >>"${sshd}"
    fi
    if grep -qE '^#?PermitRootLogin' "${sshd}"; then
      sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin prohibit-password/' "${sshd}" || true
    else
      echo 'PermitRootLogin prohibit-password' >>"${sshd}"
    fi
    if grep -qE '^#?PubkeyAuthentication' "${sshd}"; then
      sed -i 's/^#\?PubkeyAuthentication.*/PubkeyAuthentication yes/' "${sshd}" || true
    fi
    systemctl reload ssh 2>/dev/null || systemctl reload sshd 2>/dev/null || true
    info "SSH hardened (password auth disabled; root keys only)"
  fi

  systemctl enable --now fail2ban

  cat >/etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF
  dpkg-reconfigure -f noninteractive unattended-upgrades 2>/dev/null || true

  ufw --force reset >/dev/null 2>&1 || true
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow "${SSH_PORT:-22}/tcp" comment 'SSH'
  ufw --force enable
  info "UFW enabled; SSH ${SSH_PORT:-22}/tcp allowed"
}

module_hardening_uninstall() {
  # Intentionally non-destructive: do not disable UFW/SSH hardening automatically.
  info "hardening uninstall is a no-op (SSH/UFW/fail2ban left in place for safety)"
}
