# iOS Shortcut (русский)

**EN:** [../SHORTCUT-IOS.md](../SHORTCUT-IOS.md)

Панель оператора: API только из VPN + session token в файле + Face ID gate в Shortcut. Шаблон `.shortcut` на VPS **не** раздаём — собирается на устройстве по EN-инструкции.

1. Подключить VPN к FreshVPS
2. `freshvps-vpn session 72` → сохранить token в файл на iPhone
3. Shortcut: Face ID → прочитать token → HTTP к API (через VPN)
