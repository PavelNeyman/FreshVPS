#!/usr/bin/env bash
# Module: hardening
# shellcheck disable=SC2154

module_hardening_install() {
  info "Installing base hardening packages"
  pkg_install ufw fail2ban unattended-upgrades apt-listchanges needrestart \
    curl ca-certificates openssh-server

  # BBR
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

  # SSH: keys preferred; keep password auth as-is if already configured
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

  # fail2ban
  systemctl enable --now fail2ban

  # unattended-upgrades
  cat >/etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF
  dpkg-reconfigure -f noninteractive unattended-upgrades 2>/dev/null || true

  # UFW baseline
  ufw --force reset >/dev/null 2>&1 || true
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow "${SSH_PORT:-22}/tcp" comment 'SSH'
  # Other module ports will be opened by their modules after enable
  ufw --force enable
  info "UFW enabled; SSH ${SSH_PORT:-22}/tcp allowed"
}
