#!/usr/bin/env bash
# Module: metrics — prefer `netductor collect` (Go), fallback Python
# shellcheck disable=SC2154

module_metrics_install() {
  mkdir -p /var/lib/freshvps/metrics /opt/freshvps/runtime/metrics
  [[ -e /var/lib/netductor ]] || ln -sfn /var/lib/freshvps /var/lib/netductor 2>/dev/null || true

  if [[ -f "${FRESHVPS_ROOT}/runtime/metrics/collector.py" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/runtime/metrics/collector.py" /opt/freshvps/runtime/metrics/collector.py
  fi

  local exec_line="/usr/bin/python3 /opt/freshvps/runtime/metrics/collector.py"
  if [[ -x /usr/local/bin/netductor ]]; then
    exec_line="/usr/local/bin/netductor collect"
  fi

  cat >/etc/systemd/system/netductor-metrics.service <<EOF
[Unit]
Description=Netductor metrics sample
After=network-online.target

[Service]
Type=oneshot
Environment=FRESHVPS_METRICS_DIR=/var/lib/freshvps/metrics
ExecStart=${exec_line}
EOF

  cat >/etc/systemd/system/netductor-metrics.timer <<EOF
[Unit]
Description=Netductor metrics every minute

[Timer]
OnBootSec=30
OnUnitActiveSec=60
AccuracySec=10

[Install]
WantedBy=timers.target
EOF

  ln -sfn netductor-metrics.service /etc/systemd/system/freshvps-metrics.service 2>/dev/null || true
  ln -sfn netductor-metrics.timer /etc/systemd/system/freshvps-metrics.timer 2>/dev/null || true

  systemctl daemon-reload
  systemctl enable --now netductor-metrics.timer
  systemctl start netductor-metrics.service 2>/dev/null || true
  info "Metrics timer enabled (${exec_line})"
}

module_metrics_uninstall() {
  systemctl disable --now netductor-metrics.timer freshvps-metrics.timer 2>/dev/null || true
  rm -f /etc/systemd/system/netductor-metrics.{service,timer}
  rm -f /etc/systemd/system/freshvps-metrics.{service,timer}
  systemctl daemon-reload 2>/dev/null || true
}
