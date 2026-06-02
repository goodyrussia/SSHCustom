// sshcustomd — SSHCustom TPROXY-only daemon (v5.0.0)
// Single static arm64 binary. TPROXY only. No REDIRECT, no TUN, no HTTP API.
// APK controls via Unix socket + ssh.service shell script.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/GoodyOG/SSHCustom_Magisk/internal/config"
	issh "github.com/GoodyOG/SSHCustom_Magisk/internal/ssh"
	"github.com/GoodyOG/SSHCustom_Magisk/internal/proxy"
)

var version = "5.0.0"

func setLocalTimezone() {
	out, err := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output()
	if err != nil {
		return
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return
	}
	if loc, err := time.LoadLocation(name); err == nil {
		time.Local = loc
	}
}

func main() {
	setLocalTimezone()
	debug.SetGCPercent(200)
	debug.SetMemoryLimit(192 * 1024 * 1024)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "version", "-version", "--version":
		fmt.Printf("sshcustomd v%s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	case "run":
		runCmd()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  sshcustomd run -c <settings.ini> -w <workdir> [--idle]\n  sshcustomd version\n")
}

func runCmd() {
	cfgPath := flag.String("c", "", "path to settings.ini")
	workDir := flag.String("w", "", "working directory")
	idle := flag.Bool("idle", false, "start without tunnel")
	flag.Parse()

	if *cfgPath == "" || *workDir == "" {
		fmt.Fprintln(os.Stderr, "-c and -w are required")
		os.Exit(1)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("[main] config load failed: %v", err)
	}
	atomicCfg := config.NewAtomicConfig(cfg)

	st := NewState(*workDir)
	unixSrv := api.NewUnixServer(*workDir+"/run/sshcustomd.sock", st, atomicCfg)
	go unixSrv.ListenAndServe(context.Background())

	if !*idle {
		go tunnelLoop(context.Background(), atomicCfg, *cfgPath, *workDir, st)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	for {
		s := <-sig
		if s == syscall.SIGHUP {
			newCfg, err := config.Load(*cfgPath)
			if err == nil {
				atomicCfg.Store(newCfg)
			}
			continue
		}
		return
	}
}

// State holds runtime status
type State struct {
	workDir string
}

func NewState(workDir string) *State {
	return &State{workDir: workDir}
}

func tunnelLoop(ctx context.Context, atomicCfg *config.AtomicConfig, cfgPath, workDir string, st *State) {
	iptables := workDir + "/scripts/ssh.iptables"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cfg := atomicCfg.Get()
		if cfg.SSHHost == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("[tunnel] connecting to %s:%d", cfg.SSHHost, cfg.SSHPort)

		client, err := issh.Connect(cfg)
		if err != nil {
			log.Printf("[tunnel] connect failed: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Start TPROXY listeners
		tproxyCancel := startTPROXYListeners(cfg, client)

		// Apply iptables
		runScript(iptables, "enable")

		// Wait for disconnect or context cancel
		<-client.Done()
		tproxyCancel()
		runScript(iptables, "disable")
		client.Close()
	}
}

func startTPROXYListeners(cfg *config.Config, client *issh.Client) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	// SOCKS5
	go proxy.StartSOCKS5(ctx, cfg.SocksPort, client)

	// TPROXY TCP + UDP
	go proxy.StartTPROXY(ctx, cfg.TProxyPort, client)

	// DNS forwarder
	go proxy.StartDNSForwarder(ctx, cfg, client)

	return cancel
}

func runScript(script, action string) {
	cmd := exec.Command("/system/bin/sh", script, action)
	cmd.Run()
}