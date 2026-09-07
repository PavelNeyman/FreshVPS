#!/usr/bin/env bash
# Module: hardening (idempotent)
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
    # Only rewrite if not already desired (avoid churn / needless reload)
    local need_ssh=0
    grep -qE '^PasswordAuthentication[[:space:]]+no' "${sshd}" || need_ssh=1
    grep -qE '^PermitRootLogin[[:space:]]+prohibit-password' "${sshd}" || need_ssh=1
    grep -qE '^PubkeyAuthentication[[:space:]]+yes' "${sshd}" || need_ssh=1
    if [[ "${need_ssh}" -eq 1 ]]; then
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
      else
        echo 'PubkeyAuthentication yes' >>"${sshd}"
      fi
      systemctl reload ssh 2>/dev/null || systemctl reload sshd 2>/dev/null || true
      info "SSH hardened (password auth disabled; root keys only)"
    else
      info "SSH already hardened — skip"
    fi
  fi

  systemctl enable --now fail2ban 2>/dev/null || true

  if [[ ! -f /etc/apt/apt.conf.d/20auto-upgrades ]]; then
    cat >/etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF
    dpkg-reconfigure -f noninteractive unattended-upgrades 2>/dev/null || true
  fi

  # UFW: never --force reset on re-run (would drop live rules / lock out)
  if command -v ufw >/dev/null 2>&1; then
    if ufw status 2>/dev/null | grep -qi 'Status: active'; then
      ufw allow "${SSH_PORT:-22}/tcp" comment 'SSH' >/dev/null 2>&1 || true
      info "UFW already active — ensure SSH ${SSH_PORT:-22}/tcp allowed, no reset"
    else
      ufw default deny incoming
      ufw default allow outgoing
      ufw allow "${SSH_PORT:-22}/tcp" comment 'SSH'
      ufw --force enable
      info "UFW enabled; SSH ${SSH_PORT:-22}/tcp allowed"
    fi
  fi
}

module_hardening_uninstall() {
  info "hardening uninstall is a no-op (SSH/UFW/fail2ban left in place for safety)"
}
