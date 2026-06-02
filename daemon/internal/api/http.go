package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// MiniHTTPServer serves just enough HTTP for the APK's status/WAN-IP polling.
// The APK uses OkHttp to call http://127.0.0.1:9190/api/v1/status and
// /api/v1/network/public-ip. This mirrors the Unix socket getStatus output.
type MiniHTTPServer struct {
	srv     *http.Server
	getSnap func() StatusSnapshot
	getWan  func() (string, string)
}

// StartMiniHTTP creates and starts a minimal HTTP server on addr. Call Close()
// to shut it down gracefully.
func StartMiniHTTP(addr string, getSnap func() StatusSnapshot, getWan func() (string, string)) *MiniHTTPServer {
	h := &MiniHTTPServer{getSnap: getSnap, getWan: getWan}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/status", h.handleStatus)
	mux.HandleFunc("/api/v1/network/public-ip", h.handleWanIP)
	h.srv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		log.Printf("[mini-http] listening on %s", addr)
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[mini-http] %v", err)
		}
	}()
	return h
}

func (h *MiniHTTPServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return h.srv.Shutdown(ctx)
}

func (h *MiniHTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	snap := h.getSnap()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"runtime": snap,
	})
}

func (h *MiniHTTPServer) handleWanIP(w http.ResponseWriter, r *http.Request) {
	ip, country := h.getWan()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"tunnel": map[string]string{
				"ip":      ip,
				"country": country,
			},
		},
	})
}
