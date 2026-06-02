# SSHCustom APK Audit Report

**Date**: 2026-06-02  
**Source**: `/tmp/SSHCustom/app/src/main/java/com/sshcustom/app/`  
**Files audited**: 11 Kotlin files + 1 AndroidManifest.xml + 1 ProGuard file + 1 settings.ini

---

## 1. SSHControlService.kt — Shell Commands & Unix Socket Audit

**File**: `service/SSHControlService.kt` (116 lines)

### All Shell Commands
| Method | Command |
|---|---|
| `startTunnel()` | `sh $serviceScript start` → `sh /data/adb/sshcustom/scripts/ssh.service start` |
| `stopTunnel()` | `sh $serviceScript stop` → `sh /data/adb/sshcustom/scripts/ssh.service stop` |
| `restartTunnel()` | `sh $serviceScript restart` → `sh /data/adb/sshcustom/scripts/ssh.service restart` |
| `startIdle()` | `sh $serviceScript start-idle` → `sh /data/adb/sshcustom/scripts/ssh.service start-idle` |
| `isRunning()` | `sh $serviceScript status` |
| `forceCleanup()` | `sh $iptablesScript disable` → `sh /data/adb/sshcustom/scripts/ssh.iptables disable` |
| `setAutostart(true)` | `touch /data/adb/sshcustom/run/autostart` |
| `setAutostart(false)` | `rm -f /data/adb/sshcustom/run/autostart` |
| `getAutostart()` | `[ -f /data/adb/sshcustom/run/autostart ] && echo 1 \|\| echo 0` |
| `readLogFile(path, n)` | `tail -n $lines '$path' 2>/dev/null \|\| echo '(log empty)'` |
| `clearLogFile(path)` | `: > '$path'` (truncate) |
| `writeSetting(key, value)` | `sed -i 's\|^${key}=.*\|${key}="${safe}"\|' $settingsPath` |
| `writeSettings(pairs)` | Multiple `sed -i` invocations via a single shell script |
| `readSetting(key)` | `grep -E '^${key}=' $settingsPath \| head -1 \| cut -d'=' -f2- \| tr -d '"'` |

### Unix Socket Actions
**None.** The `SSHControlService` communicates with the daemon indirectly via:
- Shell commands (`ssh.service start/stop/restart/status`)
- Settings file writes (`sed -i` on `settings.ini`)

The actual Unix socket (`/data/adb/sshcustom/run/sshcustomd.sock`) is used by the **daemon** (Go binary), not by this service.

### HTTP API (port 9190) References
**None in SSHControlService.kt.** The HTTP API calls live in `Repository.kt`:
- `http://127.0.0.1:9190/api/v1/status`
- `http://127.0.0.1:9190/api/v1/network/public-ip`

### REDIRECT / TUN / tun2proxy / openvpn / vpnchain References
**None in SSHControlService.kt.** The service is clean of legacy tunnel mode references.

---

## 2. Repository.kt — Settings Keys Audit

**File**: `data/Repository.kt` (204 lines)

### SharedPreferences Keys (App-local, NOT settings.ini)
| Constant | Key | Purpose |
|---|---|---|
| `KEY_PROFILES` | `profiles_json` | Serialized profile list |
| `KEY_ACTIVE` | `active_profile_id` | Active profile UUID |
| `KEY_SETTINGS` | `app_settings_json` | Serialized AppSettings |
| `KEY_LAST_CONN` | `last_connected` | Last known daemon connected state |
| `KEY_LAST_SSHMODE` | `last_ssh_mode` | Last known SSH mode |
| `KEY_LAST_NETMODE` | `last_net_mode` | Last known network mode |
| `KEY_LAST_VER` | `last_version` | Last known daemon version |

### settings.ini Keys Written by APK (via MainViewModel)
**From `applyProfileToSettings()`:**
- `ssh_host`, `ssh_port`, `ssh_user`, `ssh_password`, `ssh_mode`
- `ssh_sni_host`, `http_proxy_host`, `http_proxy_port`
- `payload_enabled`, `payload`

**From `applySettingsToIni()`:**
- `network_mode`, `socks_port`, `tproxy_port`, `redir_port`
- `quic`, `proxy_tcp`, `proxy_udp`
- `dns_hijack_tcp`, `dns_hijack_udp`, `dns_hijack_mode`
- `channel_pool`, `channel_pool_size`, `tcp_buffer_tuning`
- `ipv6`, `hotspot_sharing`

### settings.ini Keys NOT Written by APK (present in module/settings.ini but untouched by APK)
- `cgroup_blkio` — iptables cgroup block I/O toggle
- `blkio_weight` — weight value (900)
- `module_prop` — path to module.prop

**Verdict**: The APK writes all user-configurable keys to settings.ini. The three untouched keys (`cgroup_blkio`, `blkio_weight`, `module_prop`) are module-internal/advanced config that don't need a UI toggle. ✅ Synced.

---

## 3. Models.kt — Data Class Audit

**File**: `domain/Models.kt` (74 lines)

| Data Class / Sealed Class | Fields | Used? |
|---|---|---|
| `DaemonStatus` | 14 fields (connected, uptimeSeconds, sshMode, networkMode, bytesSent, bytesRecv, channelPoolSize, channelPoolAvail, activeConnections, version, memRssMb, cpuPercent, upKbps, downKbps, lastError) | ✅ Repository, ViewModel, HomeScreen |
| `Profile` | 11 fields (id, name, host, port, user, password, mode, sniHost, proxyHost, proxyPort, payloadEnabled, payload) | ✅ Repository, ViewModel, ProfilesScreen |
| `AppSettings` | 17 fields (networkMode, socksPort, tproxyPort, redirPort, quic, proxyTcp, proxyUdp, dnsHijackTcp, dnsHijackUdp, dnsHijackMode, channelPool, channelPoolSize, tcpBufferTuning, ipv6, autostartTunnel, hotspotSharing) | ✅ Repository, ViewModel, SettingsScreen |
| `TunnelState` (sealed) | Stopped, Starting, Connected, Stopping, Error | ✅ ViewModel, HomeScreen |
| `NetSpeed` | upKbs, downKbs | ✅ Repository, ViewModel, HomeScreen |

**Verdict**: No unused data classes. ✅ All 5 types are actively used.

---

## 4. MainViewModel.kt — State Variable Audit

**File**: `MainViewModel.kt` (446 lines)

| State Variable | Type | Purpose |
|---|---|---|
| `_hasRoot` / `hasRoot` | `MutableStateFlow<Boolean>` | Root availability |
| `rootBinder` | `SSHControlService.LocalBinder?` | Bound RootService reference |
| `rootServiceBound` | `Boolean` | RootService binding status |
| `status` | `StateFlow<DaemonStatus>` | Daemon status (polled every 1s via HTTP) |
| `_tunnelStateOverride` / `tunnelState` | `StateFlow<TunnelState>` | Derived tunnel state with transition overrides |
| `_pendingAction` / `pendingAction` | `StateFlow<String?>` | In-flight action: "start"/"stop"/"restart"/null |
| `_isLoading` / `isLoading` | `StateFlow<Boolean>` | Loading state for buttons |
| `netSpeed` | `StateFlow<NetSpeed>` | Network speed (derived from daemon status) |
| `wanIp` | `StateFlow<String>` | WAN IP (polled every 60s via HTTP) |
| `_latencyGoogle` / `latencyGoogle` | `StateFlow<Int>` | Google latency (SOCKS5 TCP connect) |
| `_latencyCloudflare` / `latencyCloudflare` | `StateFlow<Int>` | Cloudflare latency (SOCKS5 TCP connect) |
| `_logText` / `logText` | `StateFlow<String>` | Log content (polled every 2s) |
| `_activeLog` / `activeLog` | `StateFlow<String>` | Active log type: "core"/"boot"/"tool" |
| `_profiles` / `profiles` | `StateFlow<List<Profile>>` | Profile list |
| `_activeProfileId` / `activeProfileId` | `StateFlow<String>` | Active profile UUID |
| `_settings` / `settings` | `StateFlow<AppSettings>` | App settings |
| `settingsNeedRestart` | `Boolean` | Whether settings changed need tunnel restart |

### vpnchain Remnants
**No vpnchain references in MainViewModel.kt.** ✅ Clean.

**HOWEVER**, `vpnchain` references exist elsewhere in the project:
- **`app/proguard-rules.pro`** (lines 6-7): `-keep class com.sshcustom.vpnchain.** { *; }` — ProGuard rules for a package that no longer exists in the app
- **`daemon/cmd/sshcustomd/main.go`** (line 63): `const pkg = "com.sshcustom.vpnchain"` — used by `whitelistApp()` for Doze/battery whitelisting. **WRONG package** — should be `com.sshcustom.app`
- **`daemon/cmd/sshcustomd/main.go`** (line 939): `const propPath = "/data/adb/modules/sshcustom-vpnchain/module.prop"` — **WRONG path** — should be `/data/adb/modules/sshcustom/module.prop` (per `module/settings.ini` line 85)

---

## 5. HomeScreen.kt — HTTP/WebUI, REDIRECT, TUN References

**File**: `ui/screens/HomeScreen.kt` (420 lines)

### HTTP/WebUI References
**None.** HomeScreen does not reference any HTTP endpoint or WebUI.

### REDIRECT / TUN Mode References
**None.** HomeScreen only displays `status.sshMode` and `status.networkMode` as display strings (line 140: `"${status.sshMode.uppercase()}  ·  ${status.networkMode.uppercase()}  ·  v${status.version}"`) but does not contain hardcoded references to "redirect", "tun", or specific modes.

### Other Checks
- HomeScreen delegates all logic to ViewModel
- UI is pure Compose, no shell commands or direct daemon interaction
- No vpnchain, tun2proxy, or openvpn references

**Verdict**: ✅ Clean.

---

## 6. SettingsScreen.kt — Toggle Audit

**File**: `ui/screens/SettingsScreen.kt` (313 lines)

| Section | Toggle / Field | settings.ini Key | Match? |
|---|---|---|---|
| Ports | Redirect Port | `redir_port` | ✅ |
| Ports | TPROXY Port | `tproxy_port` | ✅ |
| Ports | SOCKS5 Port | `socks_port` | ✅ |
| Traffic Mode | Mode dropdown (Redirect/TPROXY) | `network_mode` | ✅ |
| Proxy Behaviour | Block QUIC | `quic` (mapped: checked→"disable", unchecked→"enable") | ✅ |
| Proxy Behaviour | Proxy TCP | `proxy_tcp` | ✅ |
| Proxy Behaviour | Proxy UDP | `proxy_udp` | ✅ |
| Performance | Channel Pool | `channel_pool` | ✅ |
| Performance | TCP Buffer Tuning | `tcp_buffer_tuning` | ✅ |
| DNS Hijack | DNS Hijack TCP | `dns_hijack_tcp` | ✅ |
| DNS Hijack | DNS Hijack UDP | `dns_hijack_udp` | ✅ |
| DNS Hijack | DNS Hijack Mode | `dns_hijack_mode` | ✅ |
| IPv6 | Disable IPv6 | `ipv6` (inverted: checked→ipv6=false, unchecked→ipv6=true) | ✅ |
| Hotspot Sharing | Share tunnel via hotspot | `hotspot_sharing` | ✅ |
| Boot Behaviour | Start tunnel on boot | **NO INI KEY** — uses `/data/adb/sshcustom/run/autostart` marker file | ✅ (intentional) |
| About | Developer/Source/Version/Module data | N/A (display only) | N/A |

**Verdict**: All 15 toggles/fields have corresponding settings.ini keys except `autostartTunnel` which intentionally uses a marker file mechanism. ✅ Full coverage.

---

## 7. Package Name Verification

**Required package**: `com.sshcustom.app`

| File | Package Declaration | Match? |
|---|---|---|
| `MainActivity.kt` | `com.sshcustom.app` | ✅ |
| `MainViewModel.kt` | `com.sshcustom.app` | ✅ |
| `SSHCustomApp.kt` | `com.sshcustom.app` | ✅ |
| `SSHControlService.kt` | `com.sshcustom.app.service` | ✅ |
| `Repository.kt` | `com.sshcustom.app.data` | ✅ |
| `Models.kt` | `com.sshcustom.app.domain` | ✅ |
| `HomeScreen.kt` | `com.sshcustom.app.ui.screens` | ✅ |
| `SettingsScreen.kt` | `com.sshcustom.app.ui.screens` | ✅ |
| `LogsScreen.kt` | `com.sshcustom.app.ui.screens` | ✅ |
| `ProfilesScreen.kt` | `com.sshcustom.app.ui.screens` | ✅ |
| `Theme.kt` | `com.sshcustom.app.ui.theme` | ✅ |
| `app/build.gradle.kts` | `namespace = "com.sshcustom.app"` | ✅ |

**Verdict**: All files use correct package hierarchy. ✅

---

## 8. AndroidManifest.xml Audit

**File**: `/tmp/SSHCustom/app/src/main/AndroidManifest.xml` (37 lines)

- **No explicit `package` attribute** — modern AGP uses `namespace` from `build.gradle.kts` (confirmed: `com.sshcustom.app`)
- `android:name=".SSHCustomApp"` → resolves to `com.sshcustom.app.SSHCustomApp` ✅
- `android:name=".MainActivity"` → resolves to `com.sshcustom.app.MainActivity` ✅
- `android:name=".service.SSHControlService"` → resolves to `com.sshcustom.app.service.SSHControlService` ✅
- `android:usesCleartextTraffic="true"` — ✓ required for loopback HTTP (127.0.0.1:9190)
- Permissions: INTERNET, FOREGROUND_SERVICE, CHANGE_NETWORK_STATE ✅
- RootService runs in `android:process=":root"` ✅

**Verdict**: Manifest is correct. ✅

---

## Summary of Issues Found

### 🚨 CRITICAL: Package Mismatch in Daemon (vpnchain → app)

The daemon (`daemon/cmd/sshcustomd/main.go`) still references the legacy package `com.sshcustom.vpnchain`:

1. **Line 63** — `whitelistApp()`: `const pkg = "com.sshcustom.vpnchain"`
   - **Impact**: The daemon tries to exempt `com.sshcustom.vpnchain` from Doze/battery optimization, but the actual app package is `com.sshcustom.app`. The battery whitelisting will silently fail.

2. **Line 939** — `updateModuleProp()`: `const propPath = "/data/adb/modules/sshcustom-vpnchain/module.prop"`
   - **Impact**: The daemon updates the wrong `module.prop` path. The actual module path is `/data/adb/modules/sshcustom/module.prop` (per `module/settings.ini` line 85). Module status display in KSU/Magisk managers will always show stale/empty data.

### ⚠️ MINOR: ProGuard Rules for Removed Package

**`app/proguard-rules.pro`** (lines 6-7) keeps classes from `com.sshcustom.vpnchain.**` which no longer exists in the source. These are harmless dead rules but should be cleaned up.

### ⚠️ MINOR: settings.ini Keys Not Managed by APK

Three keys exist in `module/settings.ini` that have no UI toggle:
- `cgroup_blkio` / `blkio_weight` — iptables cgroup settings
- `module_prop` — module metadata path

These appear intentional (module-internal config), but should be documented.

### ✅ All Clear

- No HTTP API references in SSHControlService (correctly isolated in Repository)
- No REDIRECT/TUN mode hardcoding in HomeScreen
- No vpnchain remnants in Kotlin source
- All SettingsScreen toggles have matching settings.ini keys
- Package `com.sshcustom.app` is consistent across all source files and build config
- AndroidManifest.xml is correctly configured
- All 5 data classes in Models.kt are actively used
- All 16 ViewModel state variables are actively used
