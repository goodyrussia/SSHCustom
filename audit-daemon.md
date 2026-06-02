# SSHCustom Daemon Audit — v2.0.0 → v1.0.1 Rebuild

**Audit date:** 2026-06-02
**Target:** `/tmp/SSHCustom/daemon/` (8 Go files, ~2,771 lines)
**Rebuild goal:** TPROXY only, strip REDIRECT/TUN/HTTP/WebUI/SSE, arm64 static binary.

---

## 1. File-by-file line audit

### 1.1 `cmd/sshcustomd/main.go` (956 lines)

| Lines | Content | Tag | Reason |
|-------|---------|-----|--------|
| 1–7 | Package header + usage comment | **KEEP** | |
| 9–16 | Imports (minus line 17) | **KEEP** | context, json, flag, fmt, io, log, net, os/exec, os/signal, runtime, debug, strconv, strings, sync, atomic, syscall, time, tzdata, api, config, ssh, proxy |
| 17 | `"net/http"` | **STRIP** | Only used for HTTP API on :9190 |
| 18–36 | Remaining imports | **KEEP** | |
| 37–56 | `setLocalTimezone()` | **KEEP** | Core daemon init |
| 58–72 | `whitelistApp()` | **KEEP** | Battery exemption |
| 74–95 | `main()` | **KEEP** | Entry point |
| 97–99 | `usage()` | **KEEP** | |
| 101–155 | `State` struct + methods | **KEEP** | Runtime state tracking |
| 157–208 | `runCmd()` up to Unix API | **KEEP** | PID file, config load, unix socket |
| 163 | `"WebUI only"` in `--idle` help text | **REFERENCE** | Mentions WebUI — update text |
| 209–210 | `mux := buildHTTPMux(...)` call | **STRIP** | HTTP mux construction |
| 211–217 | `httpSrv := &http.Server{Addr: "127.0.0.1:9190", ...}` | **STRIP** | HTTP server on port 9190 |
| 218–223 | HTTP listen goroutine (`log.Printf("[http] listening on 127.0.0.1:9190")`) | **STRIP** | |
| 224–229 | HTTP shutdown goroutine (`httpSrv.Shutdown`) | **STRIP** | |
| 230–237 | Metrics + tunnel goroutines | **KEEP** | |
| 238–261 | Signal handling (SIGHUP reload, SIGTERM etc.) | **KEEP** | |
| 263–431 | `tunnelLoop()` | **KEEP** | Core SSH tunnel lifecycle |
| 433–439 | `nextDelay()` | **KEEP** | Backoff helper |
| 441–453 | `startListeners()` — SOCKS5 | **KEEP** | |
| 454–466 | `startListeners()` — tproxy case | **KEEP** | TPROXY listener |
| 467–478 | `startListeners()` — `default: // redirect` case | **STRIP** | Uses `TransparentServer` + `RedirPort`. Remove entire default branch; make tproxy the only path (fallthrough or explicit). |
| 480–488 | `startListeners()` — DNS forwarder | **KEEP** | |
| 490–591 | `dnsForwardLoop()` + `forwardOneDNSQuery()` | **KEEP** | UDP→TCP DNS proxy |
| 595–610 | `wanIPRefresher()` | **KEEP** | |
| 612–635 | `runScriptTimeout()`, `handleControl()` | **KEEP** | |
| 637–749 | **ENTIRE HTTP API block** | **STRIP** | `buildHTTPMux()`, all `/api/v1/*` handlers, WebUI root handler |
| 639–644 | `buildHTTPMux()` signature | **STRIP** | |
| 645–656 | `env` helper closure | **STRIP** | |
| 658–660 | `/api/v1/health` | **STRIP** | |
| 662–677 | `/api/v1/status` | **STRIP** | Also references `redir_port` |
| 679–694 | `/api/v1/control` | **STRIP** | |
| 696–716 | `/api/v1/autostart` | **STRIP** | |
| 718–722 | `/api/v1/config` | **STRIP** | |
| 727–736 | `/api/v1/network/public-ip` | **STRIP** | |
| 739–746 | WebUI root handler (`/` → `webroot/index.html`) | **STRIP** | |
| 748–749 | `return mux` | **STRIP** | |
| 751–927 | `fetchPublicIP()`, `metricsLoop()`, proc readers, helpers | **KEEP** | Core metrics |
| 929–934 | `minDuration()` | **KEEP** | |
| 936–956 | `updateModuleProp()` | **KEEP** | |

**KEEP lines in main.go:** ~790 lines
**STRIP lines in main.go:** ~150 lines (HTTP API + redirect default case)

---

### 1.2 `internal/ssh/client.go` (425 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| All | **KEEP** | Pure SSH transport — direct, SNI, SNI+HTTP-proxy modes. No HTTP/WebUI/REDIRECT/TUN references anywhere. |

---

### 1.3 `internal/proxy/socks5.go` (146 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| All | **KEEP** | Core SOCKS5 proxy listener. No HTTP/WebUI/REDIRECT/TUN references. |

---

### 1.4 `internal/proxy/transparent.go` (125 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| **All 125 lines** | **STRIP** | This is the REDIRECT proxy (`SO_ORIGINAL_DST` + iptables REDIRECT). After v1.0.1 rebuild with TPROXY-only, this entire file should be **deleted**. |

**Reasoning:** The `TransparentServer` struct uses `SO_ORIGINAL_DST` to recover the pre-NAT destination from iptables REDIRECT. In a TPROXY-only world, packets arrive with the original destination preserved as `LocalAddr()` — handled by `tproxy.go`. This file has zero use and no callers after the `default:` branch is removed from `startListeners()`.

---

### 1.5 `internal/proxy/tproxy.go` (93 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| All | **KEEP** | TPROXY listener with `IP_TRANSPARENT` socket option. Core. |

---

### 1.6 `internal/config/config.go` (190 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| 1–35 | Package, imports, struct fields through `SSHSNIHost` | **KEEP** | |
| 27–29 | `HTTPProxyHost`, `HTTPProxyPort` | **KEEP** | Used for `sni_http_proxy` SSH transport mode (not an HTTP API endpoint) |
| 31–33 | `PayloadEnabled`, `Payload` | **KEEP** | |
| 36 | `NetworkMode string // redirect \| tproxy` | **COMMENT FIX** | Change comment to `// tproxy` only |
| 37–38 | `SocksPort`, `TProxyPort` | **KEEP** | |
| 39 | `RedirPort int` | **STRIP** | Remove field — no longer used after redirect stripped |
| 41–65 | Proxy behaviour, DNS, speed, IPv6, paths | **KEEP** | |
| 67–112 | `AtomicConfig`, `NewAtomic`, `Get`, `Store`, `DefaultConfig` | **KEEP** | |
| 94 | `NetworkMode: "redirect"` | **FIX** | Change default to `"tproxy"` |
| 97 | `RedirPort: 9797` | **STRIP** | Remove from defaults |
| 117–148 | `Load()` | **KEEP** | |
| 150–190 | `apply()` | **KEEP** (mostly) | |
| 170 | `case "network_mode"` | **KEEP** | But note default should become "tproxy" |
| 173 | `case "redir_port"` | **STRIP** | Remove key handling |

---

### 1.7 `internal/dnsx/dnsx.go` (348 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| All | **KEEP** | Core DNS resolver. No HTTP/WebUI/REDIRECT/TUN references. |

**However, notable observations (see §5):**
- Upstream DNS is hardcoded to `8.8.8.8:53`
- Fallback servers hardcoded to `8.8.8.8:53`, `1.1.1.1:53`
- Cache TTL, query timeout all hardcoded constants

---

### 1.8 `internal/api/unix.go` (128 lines)

| Lines | Tag | Reason |
|-------|-----|--------|
| All | **KEEP** | Unix domain socket API for Android companion app. No HTTP/WebUI/REDIRECT/TUN references. See §3 for action audit. |

---

## 2. Complete STRIP list — every variable, function, import referencing HTTP/WebUI/REDIRECT/TUN

### 2.1 Imports to remove

| File | Line | Import | Reason |
|------|------|--------|--------|
| `cmd/sshcustomd/main.go` | 17 | `"net/http"` | Only used by buildHTTPMux/http.Server |

### 2.2 Functions to remove

| File | Lines | Function | Reason |
|------|-------|----------|--------|
| `cmd/sshcustomd/main.go` | 639–749 | `buildHTTPMux()` | Entire HTTP API + WebUI handler block |

### 2.3 Variables/constants to remove or fix

| File | Line(s) | Identifier | Issue |
|------|---------|------------|-------|
| `cmd/sshcustomd/main.go` | 163 | `"WebUI only"` in --idle help | Reference to WebUI in help text — update |
| `cmd/sshcustomd/main.go` | 210 | `mux` (local) | Remove with buildHTTPMux call |
| `cmd/sshcustomd/main.go` | 211–217 | `httpSrv` (local) | Remove http.Server instance |
| `cmd/sshcustomd/main.go` | 212 | `"127.0.0.1:9190"` | Remove port 9190 reference |
| `cmd/sshcustomd/main.go` | 467–478 | `TransparentServer` usage | Remove default redirect branch |
| `cmd/sshcustomd/main.go` | 468 | `cfg.RedirPort` | Remove reference |
| `internal/config/config.go` | 39 | `RedirPort int` | Remove struct field |
| `internal/config/config.go` | 94 | `NetworkMode: "redirect"` | Change default to `"tproxy"` |
| `internal/config/config.go` | 97 | `RedirPort: 9797` | Remove from defaults |
| `internal/config/config.go` | 173 | `"redir_port"` case | Remove from apply() switch |

### 2.4 Files to delete entirely

| File | Lines | Reason |
|------|-------|--------|
| `internal/proxy/transparent.go` | 125 | REDIRECT proxy — no callers after stripping default branch |

### 2.5 HTTP handler references to remove (all in `buildHTTPMux`, lines 639–749)

| Handler | Lines | Note |
|---------|-------|------|
| `/api/v1/health` | 658–660 | |
| `/api/v1/status` | 662–677 | Also references `redir_port` |
| `/api/v1/control` | 679–694 | |
| `/api/v1/autostart` | 696–716 | |
| `/api/v1/config` | 718–722 | |
| `/api/v1/network/public-ip` | 727–736 | |
| WebUI root (`/`) | 739–746 | Serves `webroot/index.html` |

### 2.6 No references found (clean)

These terms appear **nowhere** in the daemon codebase:
- `SSE` — zero references
- `tun` — zero references
- `tun2proxy` — zero references
- `handleStatus` — zero references (HTTP handler was inline)
- `handleConfig` — zero references
- `handleProfiles` — zero references
- `handleTunnelStart` — zero references
- `handleTunnelStop` — zero references
- `handleLatency` — zero references

---

## 3. Unix API audit (`internal/api/unix.go`)

### 3.1 Supported actions

The Unix socket API uses a JSON request/response protocol:

**Request types:**

| Type | Action | Behavior | Correct? |
|------|--------|----------|----------|
| `"ping"` | — | Returns `{"pong": "1"}` | ✅ Correct — health check |
| `"status"` | — | Returns full `StatusSnapshot` (conn state, uptime, modes, metrics) | ✅ Correct |
| `"control"` | `"start"` | Runs `ssh.service start` | ✅ Correct |
| `"control"` | `"restart"` | Runs `ssh.service restart` | ✅ Correct |
| `"control"` | `"start-idle"` | Runs `ssh.service start-idle` (daemon starts but tunnel not auto-connected) | ✅ Correct |
| `"control"` | `"stop"` | Updates module.prop to "stopped", runs `ssh.service stop` | ✅ Correct |
| `"control"` | `"reload"` | No-op — SIGHUP is delivered via signal loop | ✅ Correct (delegates to OS signal) |
| Unknown type | — | Returns error `"unknown request type"` | ✅ Correct |

### 3.2 Architecture notes

- Socket path: `<workdir>/run/sshcustomd.sock` (chmod 0666 for non-root app access)
- Uses `net.Listen("unix", ...)` with JSON encoder/decoder
- Per-connection deadline: 5 seconds
- Stateless — no session tracking needed
- All actions are appropriate; no missing actions for basic daemon management

---

## 4. Config audit (`internal/config/config.go`)

### 4.1 All keys read from settings.ini

| Key | Field | Default | Used? |
|-----|-------|---------|-------|
| `ssh_host` | `SSHHost` | `""` | ✅ Tunnel dial |
| `ssh_port` | `SSHPort` | `22` | ✅ Tunnel dial |
| `ssh_user` | `SSHUser` | `""` | ✅ SSH auth |
| `ssh_password` | `SSHPassword` | `""` | ✅ SSH auth |
| `ssh_mode` | `SSHMode` | `"direct"` | ✅ Transport selection |
| `ssh_sni_host` | `SSHSNIHost` | `""` | ✅ TLS SNI |
| `http_proxy_host` | `HTTPProxyHost` | `""` | ✅ `sni_http_proxy` mode (this is SSH transport, NOT the stripped HTTP API) |
| `http_proxy_port` | `HTTPProxyPort` | `3128` | ✅ `sni_http_proxy` mode |
| `payload_enabled` | `PayloadEnabled` | `false` | ✅ Payload injection |
| `payload` | `Payload` | `""` | ✅ Payload template |
| `network_mode` | `NetworkMode` | `"redirect"` | ⚠️ Used — default must change to `"tproxy"` |
| `socks_port` | `SocksPort` | `1080` | ✅ SOCKS5 listener |
| `tproxy_port` | `TProxyPort` | `9898` | ✅ TPROXY listener |
| `redir_port` | `RedirPort` | `9797` | ❌ **UNSUPPORTED after rebuild** — must be stripped |
| `proxy_tcp` | `ProxyTCP` | `true` | ⚠️ **Stored but unused** — no code checks this field |
| `proxy_udp` | `ProxyUDP` | `false` | ⚠️ **Stored but unused** — no code checks this field |
| `quic` | `QUIC` | `"disable"` | ⚠️ **Stored but unused** — no QUIC implementation in daemon |
| `dns_hijack_tcp` | `DNSHijackTCP` | `false` | ⚠️ **Stored but unused** — DNS forwarding is always-on, unconditional |
| `dns_hijack_udp` | `DNSHijackUDP` | `false` | ⚠️ **Stored but unused** — same as above |
| `dns_hijack_mode` | `DNSHijackMode` | `"disable"` | ⚠️ **Stored but unused** — no code checks this |
| `channel_pool` | `ChannelPool` | `true` | ⚠️ **Stored but unused** — no channel pool implemented in daemon |
| `channel_pool_size` | `ChannelPoolSize` | `8` | ⚠️ **Stored but unused** |
| `tcp_buffer_tuning` | `TCPBufferTuning` | `false` | ⚠️ **Stored but unused** |
| `ipv6` | `IPv6` | `false` | ⚠️ **Stored but unused** |
| `box_dir` | `BoxDir` | `/data/adb/sshcustom` | ⚠️ **Stored but unused** — workDir passed via CLI |
| `box_run` | `BoxRun` | `/data/adb/sshcustom/run` | ⚠️ **Stored but unused** |
| `box_log` | `BoxLog` | `.../sshcustom.log` | ⚠️ **Stored but unused** |
| `box_pid` | `BoxPID` | `.../sshcustom.pid` | ⚠️ **Stored but unused** |
| `bin_dir` | `BinDir` | `/data/adb/sshcustom/bin` | ⚠️ **Stored but unused** |

### 4.2 Summary of unsupported/dead keys

**10 keys** are parsed but never consumed by daemon code:
- `proxy_tcp`, `proxy_udp`, `quic`, `dns_hijack_tcp`, `dns_hijack_udp`, `dns_hijack_mode`, `channel_pool`, `channel_pool_size`, `tcp_buffer_tuning`, `ipv6`

**5 path keys** are parsed but the CLI `-w` flag overrides the work directory:
- `box_dir`, `box_run`, `box_log`, `box_pid`, `bin_dir`

**1 key must be stripped:**
- `redir_port` — REDIRECT mode removed

These are legacy from the shell-script-based v1.0 where `ssh.service` consumed them. In v1.0.1, the daemon ignores these but `ssh.iptables` and `ssh.service` scripts may still reference them. If the scripts are also being rebuilt, these keys are candidates for removal from `settings.ini`. If the daemon is the sole consumer, they're dead weight.

---

## 5. DNS audit (`internal/dnsx/dnsx.go` + `main.go` DNS forwarding)

### 5.1 DNS forwarding — always enabled

The `startListeners()` function (main.go:483–487) **unconditionally** starts `dnsForwardLoop`:

```go
// DNS forwarder: ssh.iptables redirects device UDP:53 to 127.0.0.1:5353,
// where this listener proxies the query as TCP DNS through the current SSH
// client to 8.8.8.8:53. Survives reconnects via curClient().
go func() {
    if err := dnsForwardLoop(ctx, "127.0.0.1:5353", "8.8.8.8:53", curClient); err != nil {
        log.Printf("[dns-forward] %v", err)
    }
}()
```

There is **no conditional** — it runs regardless of `cfg.DNSHijackTCP`, `cfg.DNSHijackUDP`, or `cfg.DNSHijackMode`. The `dns_hijack_*` config keys are parsed but never checked.

**Impact:** DNS forwarding through the SSH tunnel is always active when the tunnel is up. This is likely desirable (fail-closed DNS prevents leaks), but the config keys are misleading.

### 5.2 Hardcoded values

| Value | Location | Hardcoded to | Should be configurable? |
|-------|----------|-------------|------------------------|
| **DNS upstream** | `main.go:484` (in `startListeners`) | `"8.8.8.8:53"` | **YES** — user may want custom DNS (AdGuard, local, carrier) |
| **DNS listen address** | `main.go:484` (in `startListeners`) | `"127.0.0.1:5353"` | Maybe — tied to iptables rules; change only if iptables rules change |
| **Fallback DNS servers** | `dnsx.go:41` | `["8.8.8.8:53", "1.1.1.1:53"]` | **YES** — should respect user DNS preference |
| **Cache TTL** | `dnsx.go:28` | `5 * time.Minute` | Maybe — reasonable default, but configurable for low-TTL environments |
| **Query timeout** | `dnsx.go:29` | `5 * time.Second` | Maybe — reasonable default |
| **DNS server list cache age** | `dnsx.go:30` | `30 * time.Second` | No — internal optimization |
| **TCP keepalive** | `dnsx.go:34` | `15 * time.Second` | No — internal optimization |

### 5.3 dnsx.go architecture

The `dnsx` package is used by `internal/ssh/client.go` for resolving the SSH server hostname (via `dnsx.New().DialContext()`). It:

1. Reads carrier DNS servers from `getprop` (Android properties)
2. Falls back to hardcoded public DNS (`8.8.8.8`, `1.1.1.1`)
3. Performs raw UDP DNS A-record queries (no stub resolver dependency)
4. Caches results for 5 minutes in-process
5. Sets `SO_KEEPALIVE` on dialed sockets

This is well-designed for Android (no `/etc/resolv.conf`, no CGO). The only concern is the hardcoded fallback servers.

---

## 6. Additional findings

### 6.1 `proxy_tcp`/`proxy_udp` dead fields

Config fields `ProxyTCP` and `ProxyUDP` are parsed but **never checked**. The daemon always proxies all TCP traffic via TPROXY. UDP proxying is not implemented — `proxy_udp` default is `false` and there is no UDP proxy listener. These should be removed or implemented.

### 6.2 `QUIC` dead field

`QUIC` config key (`disable`/`enable`) is parsed but never used. No QUIC implementation exists in the daemon.

### 6.3 `channel_pool` dead fields

`ChannelPool` and `ChannelPoolSize` are parsed but no SSH channel pool is implemented. The daemon creates a new SSH channel per connection via `DialTCP`. Pooling would benefit high-concurrency scenarios.

### 6.4 `--idle` flag

The `--idle` flag still works functionally (skips tunnel startup), but its help text says "WebUI only" — after stripping the HTTP/WebUI, this text should be updated to something like "start in idle mode (no auto-connect)".

### 6.5 `network_mode` default

`DefaultConfig()` sets `NetworkMode: "redirect"` as the fallback. After stripping REDIRECT, this must be changed to `"tproxy"` or the daemon will have no transparent proxy listener when `network_mode` is unset.

### 6.6 `startListeners` redirect fallthrough

```go
switch cfg.NetworkMode {
case "tproxy":
    // ... TPROXY listener
default: // redirect
    // ... REDIRECT listener via TransparentServer
}
```

After stripping, the `default:` case should either panic/error (unknown mode) or default to tproxy. Best practice: make `"tproxy"` the only valid value and reject anything else.

### 6.7 File to delete

`internal/proxy/transparent.go` — 125 lines. After removing the default branch in `startListeners`, this file has zero callers and can be safely deleted.

---

## 7. Summary of required changes for v1.0.1 rebuild

### Files to modify

| File | Change | Lines affected |
|------|--------|---------------|
| `cmd/sshcustomd/main.go` | Remove `"net/http"` import | 1 line |
| `cmd/sshcustomd/main.go` | Update `--idle` help text (remove "WebUI only") | 1 line |
| `cmd/sshcustomd/main.go` | Remove HTTP server block (lines 209–229) | ~21 lines |
| `cmd/sshcustomd/main.go` | Remove `buildHTTPMux()` function (lines 637–749) | ~113 lines |
| `cmd/sshcustomd/main.go` | Remove `default: // redirect` branch (lines 467–478) | ~12 lines |
| `internal/config/config.go` | Remove `RedirPort` field | 1 line |
| `internal/config/config.go` | Change `NetworkMode` default to `"tproxy"` | 1 line |
| `internal/config/config.go` | Remove `RedirPort` from `DefaultConfig()` | 1 line |
| `internal/config/config.go` | Remove `"redir_port"` case from `apply()` | 3 lines |
| `internal/config/config.go` | Update `NetworkMode` comment | 1 line |

### Files to delete

| File | Lines |
|------|-------|
| `internal/proxy/transparent.go` | 125 lines |

### Optional improvements (not blocking)

1. Make DNS upstream configurable via `settings.ini` key (e.g., `dns_upstream`)
2. Make DNS fallback servers configurable
3. Remove or implement dead config keys (`proxy_tcp`, `proxy_udp`, `quic`, `channel_pool`, `channel_pool_size`, `tcp_buffer_tuning`, `dns_hijack_tcp`, `dns_hijack_udp`, `dns_hijack_mode`, `ipv6`)
4. Make `network_mode` validation strict (reject anything other than `"tproxy"`)

### Estimated net change

| Metric | Before | After |
|--------|--------|-------|
| Total files | 8 | 7 |
| Total lines | ~2,771 | ~2,481 |
| Lines removed | — | ~290 |
| Go imports | 37 | 36 |

---

*Audit completed by Hermes Agent — no files were modified; findings only.*
