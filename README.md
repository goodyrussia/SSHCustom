# SSHCustom

One-click transparent SSH tunnel VPN for Magisk/KernelSU. Entire system traffic through SSH. No VpnService. Any rooted Android, any ROM.

## Architecture

```
kernel TPROXY → hev-socks5-tproxy → SOCKS5 → sshcustomd → SSH server
                                              ↓
                                         WebUI :9190
                                         Companion app
```

## Features

- **One-click transparent proxy** — entire device traffic through SSH
- **TPROXY** — full TCP+UDP support, DNS through tunnel
- **REDIRECT fallback** — works on any kernel
- **Companion app** — Material You Android app with profiles, logs, settings
- **Battery efficient** — no VpnService overhead
- **Static binaries** — zero dependencies

## Quick Start

1. Flash the module in Magisk/KernelSU
2. Reboot
3. Install the APK
4. Open app → grant root → configure SSH → Start

## Credits

Built with:
- [hev-socks5-tproxy](https://github.com/heiher/hev-socks5-tproxy) — TPROXY bridge (MIT)
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) — SSH library

## License

Apache 2.0
