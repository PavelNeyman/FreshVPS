# Установка FreshVPS (русский)

Готовая система: VPN (VLESS+HY2), Blocky, панели на localhost, CLI/Telegram.

## VPS (Debian)

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
# /etc/freshvps/READY.txt
sudo freshvps-doctor
```

Неинтерактивно: `sudo bash install.sh --config configs/freshvps.conf.example --non-interactive`

## OpenWrt (на роутере)

```bash
scp -r openwrt root@ROUTER:/root/freshvps-openwrt
ssh root@ROUTER
cd /root/freshvps-openwrt
cp site.conf.example site.conf   # пароль Wi-Fi, LAN; WAN_DEVICE=auto или eth0.2
sh install-openwrt.sh
```

## Edge SBC

`sudo bash install.sh --role edge-client --config configs/edge-client.conf.example`

Английская версия: [../INSTALL.md](../INSTALL.md)
