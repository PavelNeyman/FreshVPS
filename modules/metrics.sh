#!/usr/bin/env bash
# Module: lightweight metrics collector (no Prometheus by default)
# shellcheck disable=SC2154

module_metrics_install() {
  mkdir -p /var/lib/freshvps/metrics /opt/freshvps/runtime/metrics
  install -m 755 "${FRESHVPS_ROOT}/runtime/metrics/collector.py" /opt/freshvps/runtime/metrics/collector.py

  cat >/etc/systemd/system/freshvps-metrics.service <<EOF
[Unit]
Description=FreshVPS metrics sample
After=network-online.target

[Service]
Type=oneshot
Environment=FRESHVPS_API_ROOT=/opt/freshvps/runtime/api
Environment=FRESHVPS_METRICS_DIR=/var/lib/freshvps/metrics
ExecStart=/usr/bin/python3 /opt/freshvps/runtime/metrics/collector.py
EOF

  cat >/etc/systemd/system/freshvps-metrics.timer <<EOF
[Unit]
Description=FreshVPS metrics every minute

[Timer]
OnBootSec=30
OnUnitActiveSec=60
AccuracySec=10

[Install]
WantedBy=timers.target
EOF

  systemctl daemon-reload
  systemctl enable --now freshvps-metrics.timer
  # one sample now
  /usr/bin/python3 /opt/freshvps/runtime/metrics/collector.py 2>/dev/null || true
  info "Metrics timer enabled (1/min → /var/lib/freshvps/metrics/history.jsonl)"
}

module_metrics_uninstall() {
  systemctl disable --now freshvps-metrics.timer 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-metrics.service /etc/systemd/system/freshvps-metrics.timer
  systemctl daemon-reload 2>/dev/null || true
  info "Metrics timer removed (history kept unless purge)"
}
