#!/system/bin/sh
# uninstall.sh — Clean removal of SSHCustom.
# Called by Magisk/KernelSU/APatch when the module is uninstalled.
# Stops any running services, removes iptables rules, and deletes all
# module data from /data/adb/sshcustom/ so nothing is left behind.

MODDIR="${0%/*}"
WORK_DIR="/data/adb/sshcustom"
SERVICE="${WORK_DIR}/scripts/ssh.service"
IPTABLES="${WORK_DIR}/scripts/ssh.iptables"


# Stop SSH tunnel + daemon
[ -x "${SERVICE}" ] && sh "${SERVICE}" stop >/dev/null 2>&1 || true

# Clean iptables (in case stop didn't fully clear)
[ -x "${IPTABLES}" ] && sh "${IPTABLES}" disable >/dev/null 2>&1 || true

# Kill any remaining daemon processes
killall sshcustomd 2>/dev/null || true

# Restore IPv6 (in case it was disabled)
sysctl -w net.ipv6.conf.all.disable_ipv6=0 >/dev/null 2>&1 || true
sysctl -w net.ipv6.conf.default.disable_ipv6=0 >/dev/null 2>&1 || true

# Remove ALL module data
rm -rf "${WORK_DIR}"

# Remove battery optimization whitelist for the companion app
dumpsys deviceidle whitelist -com.sshcustom.app >/dev/null 2>&1 || true

exit 0
