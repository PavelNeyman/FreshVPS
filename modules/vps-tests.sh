#!/usr/bin/env bash
# Install VPS test runner (always available on host)
# shellcheck disable=SC2154

module_vps_tests_install() {
  mkdir -p /opt/freshvps/bin /opt/freshvps/scripts
  install -m 755 "${FRESHVPS_ROOT}/scripts/vps-tests.sh" /opt/freshvps/scripts/vps-tests.sh
  cat >/opt/freshvps/bin/freshvps-tests <<'EOF'
#!/usr/bin/env bash
exec bash /opt/freshvps/scripts/vps-tests.sh "$@"
EOF
  chmod 755 /opt/freshvps/bin/freshvps-tests
  ln -sfn /opt/freshvps/bin/freshvps-tests /usr/local/bin/freshvps-tests
  info "VPS tests: freshvps-tests | freshvps-tests --list | freshvps-tests --default"
}

module_vps_tests_uninstall() {
  rm -f /usr/local/bin/freshvps-tests /opt/freshvps/bin/freshvps-tests
}
