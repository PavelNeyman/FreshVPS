# План тестов ИИ (на VPS по SSH)

**EN:** [../TEST-PLAN-AI.md](../TEST-PLAN-AI.md)

Запускать **после** того, как владелец дал SSH на чистую/переустановленную Debian VPS. Без физических клиентов, UI Telegram и «ощущения» Wi-Fi.

## Предусловия

- Root SSH по ключу
- Исходящий HTTPS к GitHub / codeload
- Предпочтительно Debian 12+

## Фаза A — Bootstrap

1. Предпочтительно файл:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
   sudo bash /tmp/fv.sh --non-interactive
   ```
   Или с `/root/freshvps.conf`, если владелец подготовил.
2. Полный лог. Ожидаем exit 0 или только задокументированные WARN.
3. `test -f /var/lib/freshvps/installed_version` → версия ≥ 0.4.1.
4. `test -f /etc/freshvps/install.conf`.
5. `test -f /etc/freshvps/clients/operator/subscription.txt` и не пустой.

## Фаза B — Инструменты хоста

6. `command -v freshvps-doctor freshvps-vpn freshvps-tests freshvps-smoke`
7. `sudo freshvps-doctor` → exit 0; без FAIL; WARN (variant) допустим.
8. `sudo freshvps-smoke` → exit 0.
9. `sudo freshvps-vpn list` → есть `operator` on.
10. `sudo freshvps-vpn link operator` → subscription (vless + hysteria2).

## Фаза C — Сервисы

11. `systemctl is-active sing-box blocky`
12. `sing-box check -c /usr/local/etc/sing-box/config.json`
13. `dig @127.0.0.1 example.com +short` не пусто
14. `ss -lntp | grep -E ':443|:8443|:53'` (или аналог)
15. Если API включён: `curl -fsS http://127.0.0.1:8787/health`
16. Session: `TOK=$(sudo freshvps-vpn session 1 | head -1)` затем  
    `curl -fsS -H "Authorization: Bearer $TOK" http://127.0.0.1:8787/vpn/users`
17. Docker-панели (если включены): `docker ps`; порты на 127.0.0.1.

## Фаза D — Идемпотентный upgrade

18. Повтор:
    ```bash
    sudo bash /tmp/fv.sh --upgrade --non-interactive
    ```
19. Ожидаем skip здоровых модулей; UFW active; UUID operator тот же (jq).
20. `freshvps-doctor` снова exit 0.

## Фаза E — Жизненный цикл пользователя (только сервер)

21. `freshvps-vpn add aitest`
22. Артефакты в `/etc/freshvps/clients/aitest/subscription.txt`
23. `disable` / `enable` / `revoke aitest`
24. После revoke каталог удалён; sing-box active.

## Фаза F — Негатив / safety

25. `PasswordAuthentication no` **только если** есть authorized_keys.
26. SSH жив (UFW не «убит» reset).
27. API без Bearer → 401.
28. Опционально: `freshvps-tests sysbench`; полный `--default` при наличии времени.

## Вне скоупа ИИ

Телефон/TV, живой Wi-Fi, UX Telegram, качество стрима, кабинет провайдера.

## Критерий pass

Фазы A–E зелёные, нет lockout SSH, doctor fail=0. Ошибки фиксировать с командой и stderr.
