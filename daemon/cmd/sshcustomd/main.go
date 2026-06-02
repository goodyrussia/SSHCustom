// sshcustomd v5.0.0 - TPROXY only SSH VPN daemon
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
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/goodyrussia/SSHCustom/daemon/internal/config"
	"github.com/goodyrussia/SSHCustom/daemon/internal/proxy"
)

var version = "5.0.0"

func setLocalTimezone() {
	out, _ := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output()
	if name := strings.TrimSpace(string(out)); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			time.Local = loc
		}
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
	case "version":
		fmt.Printf("sshcustomd v%s\n", version)
	case "run":
		runCmd()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "sshcustomd run -c <settings.ini> -w <workdir>")
}

func runCmd() {
	cfgPath := flag.String("c", "", "settings.ini")
	workDir := flag.String("w", "", "workdir")
	idle := flag.Bool("idle", false, "idle mode")
	flag.Parse()

	if *cfgPath == "" || *workDir == "" {
		fmt.Fprintln(os.Stderr, "-c and -w required")
		os.Exit(1)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	atomicCfg := config.NewAtomicConfig(cfg)

	// Unix socket API for APK
	go startUnixAPI(*workDir)

	if !*idle {
		go tunnelLoop(atomicCfg, *workDir)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}

func tunnelLoop(atomicCfg *config.AtomicConfig, workDir string) {
	iptables := workDir + "/scripts/ssh.iptables"
	for {
		cfg := atomicCfg.Get()
		if cfg.SSHHost == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("[tunnel] connecting %s", cfg.SSHHost)

		// TODO: real SSH client + payload injection
		// For now just start listeners

		cancel := startListeners(cfg)

		runScript(iptables, "enable")

		time.Sleep(30 * time.Second) // placeholder until real SSH done

		cancel()
		runScript(iptables, "disable")
	}
}

func startListeners(cfg *config.Config) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go proxy.StartSOCKS5(ctx, cfg.SocksPort, nil)
	go proxy.StartTPROXY(ctx, cfg.TProxyPort, nil)
	return cancel
}

func runScript(script, action string) {
	exec.Command("/system/bin/sh", script, action).Run()
}

func startUnixAPI(workDir string) {
	// Unix socket server for APK control (ping/status/control)
	log.Println("[unix] listening on", workDir+"/run/sshcustomd.sock")
}