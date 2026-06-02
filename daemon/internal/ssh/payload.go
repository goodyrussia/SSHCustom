package ssh

import "strings"

// InjectPayload applies HTTP Injector style payload to the request
func InjectPayload(host, payload string, mode string) string {
	if payload == "" {
		return host
	}
	switch mode {
	case "front", "query":
		return payload + host
	case "back":
		return host + payload
	case "split":
		parts := strings.Split(host, ".")
		if len(parts) > 1 {
			return parts[0] + payload + "." + strings.Join(parts[1:], ".")
		}
		return host + payload
	default:
		return host
	}
}