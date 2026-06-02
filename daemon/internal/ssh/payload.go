package ssh

import "strings"

// Payload modes (HTTP Injector style)
const (
	ModeFront  = "front"
	ModeBack   = "back"
	ModeQuery  = "query"
	ModeDual   = "dual"
	ModeSplit  = "split"
)

// Inject applies payload based on mode
func Inject(host, payload, mode string) string {
	if payload == "" || mode == "" {
		return host
	}
	switch mode {
	case ModeFront, ModeQuery:
		return payload + host
	case ModeBack:
		return host + payload
	case ModeDual:
		return payload + host + payload
	case ModeSplit:
		if idx := strings.Index(host, "."); idx > 0 {
			return host[:idx] + payload + host[idx:]
		}
		return host + payload
	default:
		return host
	}
}