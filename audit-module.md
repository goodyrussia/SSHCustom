# SSHCustom Module Audit — Complete Line-by-Line

**Date:** 2026-06-02
**Module version in code:** v2.3.13 (module.prop line 3)
**Target:** v1.0.1 rebuild — TPROXY ONLY, strip REDIRECT/TUN

---

## 1. File Inventory

| File | Lines | Size | Purpose |
|------|-------|------|---------|
| `scripts/ssh.iptables` | 645 | 27,863 B | Traffic interception rules |
| `scripts/ssh.tool` | 176 | 7,073 B | Helper utilities |
| `scripts/ssh.service` | 226 | 8,054 B | Service lifecycle manager |
| `customize.sh` | 111 | 5,047 B | Magisk/KSU/APatch install |
| `service.sh` | 84 | 2,878 B | Boot entrypoint |
| `action.sh` | 37 | 955 B | KSU/APatch action button |
| `uninstall.sh` | 33 | 1,073 B | Clean removal |
| `settings.ini` | 102 | 4,200 B | Configuration |
| `module.prop` | 7 | 114 B | Module metadata |

---

## 2. ssh.iptables — Detailed Audit

### 2.1 REDIRECT — Everything tagged STRIP

| Line(s) | Symbol | Action |
|---------|--------|--------|
| 268–316 | `start_redirect()` | **STRIP** — entire function |
| 318–332 | `stop_redirect()` | **STRIP** — entire function |
| 521–527 | `apply_redirect_family()` | **STRIP** — entire function |
| 239–253 | `setup_dns_hijack()` | **STRIP** — uses REDIRECT in nat table; controlled by `dns_hijack_mode` |
| 256–262 | `cleanup_dns_hijack()` | **STRIP** — cleans SSHC_DNS_HIJACK{,6} chains |
| 572–580 | Main `redirect` case | **STRIP** — lines 572–580 in case statement |
| 534–535 | TPROXY→redirect fallback | **STRIP** — in `apply_tproxy_family()`, fallback to redirect |
| 122–129 | `apply_capability_downgrades()` redirect fallback | **STRIP** — downgrades to redirect; change to abort/fail instead |
| 296–310 | Hotspot in `start_redirect()` | **STRIP** — uses REDIRECT in nat table; TPROXY hotspot needed if feature kept |

**REDIRECT chains to strip (all in `nat` table):**
- `SSHC_EXTERNAL`, `SSHC_LOCAL`, `SSHC_EXTERNAL6`, `SSHC_LOCAL6`
- `SSHC_HOTSPOT`, `SSHC_HOTSPOT6`
- `SSHC_DNS_HIJACK`, `SSHC_DNS_HIJACK6`

### 2.2 TUN — Everything tagged STRIP

| Line(s) | Symbol | Action |
|---------|--------|--------|
| 29–33 | TUN constants (`tun_bypass_mark`, `tun_bypass_pref`, `tun_route_mark`, `tun_route_table`, `tun_route_pref`) | **STRIP** |
| 405–456 | `start_tun_bypass()` | **STRIP** — entire function |
| 458–482 | `stop_tun_bypass()` | **STRIP** — entire function |
| 539–545 | `apply_tun_family()` | **STRIP** — entire function |
| 591–599 | Main `tun|tun_udpgw` case | **STRIP** — lines 591–599 in case statement |

**TUN chains to strip (all in `mangle` table):**
- `SSHC_TUN_PRE`, `SSHC_TUN_OUT`, `SSHC_TUN_PRE6`, `SSHC_TUN_OUT6`

### 2.3 `cleanup_all()` Impacted (line 486–493)

Must remove:
- Line 487: `stop_redirect    2>/dev/null || true`
- Line 489: `stop_tun_bypass  2>/dev/null || true`

Keep:
- Line 488: `stop_tproxy      2>/dev/null || true`
- Line 490: `cleanup_quic_block 2>/dev/null || true`
- Line 491: `cleanup_dns_hijack 2>/dev/null || true` — **STRIP** (DNS hijack is REDIRECT-based)
- Line 492: `cleanup_dns_forward 2>/dev/null || true`

### 2.4 TPROXY — Verified Present

| Function | Lines | Status |
|----------|-------|--------|
| `start_tproxy()` | 337–383 | ✅ Present — mangle table, SSHC_TP_EXT/LOC chains, TPROXY target with fwmark |
| `stop_tproxy()` | 385–399 | ✅ Present — cleans chains + policy routes |
| `apply_tproxy_family()` | 529–537 | ✅ Present — calls start_tproxy |
| `probe_capabilities()` | 77–116 | ✅ Present — checks `/proc/net/ip_tables_targets` for TPROXY |
| `apply_capability_downgrades()` | 118–133 | ⚠️ Needs update — currently falls back to REDIRECT; should abort |
| Captive portal disable | 558–563 | ✅ Present in `enable|renew` case |
| Captive portal restore | 625–630 | ✅ Present in `disable` case |

### 2.5 DNS Handling

| Function | Lines | Strategy |
|----------|-------|----------|
| `setup_dns_forward()` | 142–150 | ✅ **KEEP** — DNAT UDP:53 → 127.0.0.1:5353; skips uid 0; always-on |
| `cleanup_dns_forward()` | 152–156 | ✅ **KEEP** — reverses above |
| `setup_dns_hijack()` | 239–253 | ❌ **STRIP** — REDIRECT-based, separate from dns_forward |
| `cleanup_dns_hijack()` | 256–262 | ❌ **STRIP** |

**Note:** DNS-through-tunnel (`setup_dns_forward`) is hardcoded always-on — no `[dns_forward]` section exists in settings.ini. This matches the v1.0.1 goal ("DNS through tunnel always on").

### 2.6 `-w 100` Coverage

**Finding: CORRECT and CONSISTENT**

Lines 14–22 set `IPV`/`IP6V` with `-w 100` for Android 11+:
```sh
if [ "${_build_major:-0}" -ge 11 ] 2>/dev/null; then
  IPV="iptables -w 100"
  IP6V="ip6tables -w 100"
else
  IPV="iptables"
  IP6V="ip6tables"
fi
```

All 88 iptables invocations use either `${IPV}`, `${IP6V}`, or `${iptables}` (set to one of those at call time). **Zero bare iptables/ip6tables calls exist.** ✅

### 2.7 tun2proxy / openvpn / vpnchain References

**Finding: NONE** — Zero references found across all module files. ✅

---

## 3. ssh.tool — Audit

### 3.1 speed_boost Sysctls

| Function | Lines | Details |
|----------|-------|---------|
| `apply_speed_boost()` | 117–141 | Sets: `net.core.rmem_max=134217728`, `net.core.wmem_max=134217728`, `tcp_rmem="4096 87380 134217728"`, `tcp_wmem="4096 65536 134217728"` |
| `restore_speed_settings()` | 145–161 | Restores from `${box_run}/speed_orig.env` |

**⚠️ Minor Note:** Function is named "speed_boost" but only does TCP buffer tuning. No BBR sysctl (`net.ipv4.tcp_congestion_control=bbr`) is applied. Not a bug — just naming.

### 3.2 Utility Functions

| Function | Lines | Status |
|----------|-------|--------|
| `cgroup_blkio()` | 62–114 | ✅ Assigns daemon to blkio cgroup; safely handles unwritable cgroups |
| `probe_user_group()` | 43–59 | ✅ Returns `user:group` of running daemon |
| `current_boot_id()` | 32–39 | ✅ Strips hyphens from boot_id |
| Main dispatcher | 164–175 | ✅ 5 commands: blkio, speed_boost, speed_restore, probe_user, boot_id |

---

## 4. ssh.service — Audit

### 4.1 Lifecycle

| Command | Lines | Status |
|---------|-------|--------|
| `start` | 208 | ✅ `start_service()` — starts daemon, applies blkio |
| `start-idle` | 209–213 | ✅ Daemon with `--idle` flag (WebUI only) |
| `stop` | 214 | ✅ `stop_service()` — disables iptables, kills daemon, restores speed |
| `restart` | 215–219 | ✅ stop + sleep 2 + start |
| `status` | 220 | ✅ Checks PID, RSS, API |

### 4.2 Key Behaviors

- **Line 97:** Speed boost applied before daemon start ✅
- **Lines 147–151:** Commentary says iptables are "now managed by the daemon" — daemon applies rules only after tunnel connects ✅
- **Lines 182:** Speed settings restored on stop ✅
- **Line 163:** `sh "${IPTABLES}" disable` called on stop — loads runtime snapshot for targeted cleanup ✅

---

## 5. Other Scripts

### 5.1 customize.sh

| Feature | Lines | Status |
|---------|-------|--------|
| ABI check | 11–19 | ✅ arm64-v8a only; aborts on other archs |
| Binary install path | 47 | ✅ `${MODPATH}/bin/arm64-v8a/$1` |
| Settings preserve | 78–88 | ✅ Preserves existing settings.ini; no merge yet |
| Verification | 92–104 | ✅ Verifies all critical files installed |

**Note:** `bin/armeabi-v7a/sshcustomd` exists but is never installed (only arm64 path used). Could be cleaned up.

### 5.2 service.sh

| Feature | Lines | Status |
|---------|-------|--------|
| Wait for boot | 20–28 | ✅ Waits for `sys.boot_completed=1`, max 5 min (100 × 3s) |
| Daemon start | 33–37 | ✅ Starts idle mode on boot |
| Module.prop update | 37 | ✅ Sets description to `[ 💤 ] SSHCustom idle` |
| Autostart | 44–80 | ✅ Checks marker, waits for route (60s), API (10s), then starts |
| Logging | 15–81 | ✅ All output → `${RUN_DIR}/boot.log` |

### 5.3 uninstall.sh

✅ Stops service → cleans iptables → kills daemon → restores IPv6 → removes `/data/adb/sshcustom/`

### 5.4 action.sh

✅ Toggle (default), start, stop, status — delegates to ssh.service

---

## 6. settings.ini — Unused Keys After REDIRECT Strip

### Keys to REMOVE (only used by REDIRECT path):

| Key | Line | Used In |
|-----|------|---------|
| `redir_port="9799"` | 52 | `start_redirect()`, `setup_dns_hijack()` |
| `dns_hijack_tcp="false"` | 63 | `setup_dns_hijack()`, `start_tun_bypass()` (TUN) |
| `dns_hijack_udp="false"` | 64 | `setup_dns_hijack()`, `start_tun_bypass()` (TUN) |
| `dns_hijack_mode="disable"` | 65 | `setup_dns_hijack()`, `runtime_save()`, `apply_capability_downgrades()` |

### Keys to KEEP (still used):

| Key | Used In |
|-----|---------|
| `network_mode` | Main dispatch, `apply_capability_downgrades()` — will be fixed to `tproxy` only |
| `tproxy_port` | `start_tproxy()`, `runtime_save()` |
| `socks_port` | Daemon (runtime_save only in iptables, actual use in daemon) |
| `proxy_tcp`, `proxy_udp` | `start_tproxy()` |
| `quic` | `apply_quic_block()` |
| `hotspot_sharing` | ⚠️ Currently only in REDIRECT mode. Needs TPROXY implementation or removal. |
| `tcp_buffer_tuning` | `apply_speed_boost()` |
| `channel_pool`, `channel_pool_size` | Daemon |
| `cgroup_blkio`, `blkio_weight` | `cgroup_blkio()` |
| `ipv6` | `enable_ipv6()`, `disable_ipv6()` |

### No `[dns_forward]` Section

DNS forwarding through tunnel is **hardcoded on** via `setup_dns_forward()` in `enable|renew` case (line 611). There is no corresponding settings key. This is **correct per v1.0.1 spec** — just confirm intent (always-on vs. configurable).

---

## 7. module.prop — Needs Update

```ini
# Current
version=v2.3.13
versionCode=231300

# Target
version=v1.0.1
versionCode=100100
```

Also update `description` — currently `[ 💤 ] SSHCustom idle` (set by service.sh at boot).

---

## 8. Summary of Required Changes for v1.0.1

### ssh.iptables — Lines to strip:

1. **L29–33:** TUN constants
2. **L122–129:** Redirect downgrade in `apply_capability_downgrades()` — change to log+abort
3. **L239–253:** `setup_dns_hijack()` entire function
4. **L256–262:** `cleanup_dns_hijack()` entire function
5. **L268–316:** `start_redirect()` entire function
6. **L318–332:** `stop_redirect()` entire function
7. **L405–456:** `start_tun_bypass()` entire function
8. **L458–482:** `stop_tun_bypass()` entire function
9. **L487:** Remove `stop_redirect` from `cleanup_all()`
10. **L489:** Remove `stop_tun_bypass` from `cleanup_all()`
11. **L491:** Remove `cleanup_dns_hijack` from `cleanup_all()`
12. **L506:** Remove `redir_port` from `runtime_save()`
13. **L508:** Remove `tun_device` from `runtime_save()`
14. **L521–527:** `apply_redirect_family()` entire function
15. **L529–537:** Update `apply_tproxy_family()` — remove redirect fallback (lines 533–536)
16. **L539–545:** `apply_tun_family()` entire function
17. **L572–580:** Redirect case in main dispatch
18. **L591–599:** TUN case in main dispatch

### ssh.tool — No changes needed
### ssh.service — No changes needed
### customize.sh — No changes needed (arm64-v8a path correct)
### service.sh — No changes needed
### action.sh — No changes needed
### uninstall.sh — No changes needed

### settings.ini — Remove keys:
- `redir_port`
- `dns_hijack_tcp`
- `dns_hijack_udp`
- `dns_hijack_mode`
- Update `network_mode` comment to `tproxy` only

### module.prop — Update:
- `version=v1.0.1`
- `versionCode=100100`

---

## 9. Open Questions

1. **Hotspot sharing:** Currently only implemented in REDIRECT mode. Does v1.0.1 need TPROXY hotspot? If not, remove `hotspot_sharing` key.
2. **TPROXY capability failure:** Current fallback is redirect. v1.0.1 should abort with clear error.
3. **DNS forward toggle:** Always-on is hardcoded. Confirm this is the intended design (vs. adding a `dns_forward=true/false` key).
4. **armeabi-v7a binary:** Exists in `bin/armeabi-v7a/` but never installed. Clean up?
