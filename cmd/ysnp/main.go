package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/virtualpeter/ysnp/internal/server"
)

func main() {
	cfg, listen, logFlags, configPath := parseFlags()
	configureLog(logFlags)

	if !server.ValidStatus(cfg.RedirectStatus) {
		slog.Error("redirect status must be one of 301, 302, 307, 308")
		os.Exit(1)
	}

	routes, err := server.LoadRoutes(configPath)
	if err != nil {
		slog.Error("load config", "path", configPath, "err", err)
		os.Exit(1)
	}
	cfg.Routes = routes

	if err := cfg.Validate(); err != nil {
		slog.Error("configuration", "err", err)
		os.Exit(1)
	}

	slog.Info("configuration",
		"listenAddr", listen,
		"targetProto", cfg.TargetProto,
		"targetHost", cfg.TargetHost,
		"targetPort", cfg.TargetPort,
		"targetPath", cfg.TargetPath,
		"allowedHosts", cfg.AllowedHosts,
		"config", configPath,
		"routes", len(cfg.Routes),
	)

	mux := http.NewServeMux()
	mux.Handle("/", server.Handler(cfg))

	httpSrv := &http.Server{
		Addr:    listen,
		Handler: server.Logger(cfg, mux),
	}
	slog.Error("server stopped", "err", httpSrv.ListenAndServe())
	os.Exit(1)
}

func parseFlags() (server.Config, string, string, string) {
	listen := flag.String("listen", defaultListen(), "TCP host:port to listen on for http requests")
	targetProto := flag.String("target_proto", env("TARGET_PROTO", "https"), "protocol to redirect to, so far the only other supported option is http")
	targetHost := flag.String("target_host", env("TARGET_HOST", ""), "hardcode this domainname in redirect instead of passing on request")
	targetPort := flag.String("target_port", env("TARGET_PORT", ""), "port to use in redirect, default is to not have an explicit port")
	targetPath := flag.String("target_path", env("TARGET_PATH", ""), "hardcode this path in redirect, default means use request path")
	blockQuery := flag.Bool("blockquery", envBool("BLOCKQUERY", false), "set if you want to block passing of request query parameters in redirect")
	redirectStatus := flag.Int("status", envInt("STATUS", http.StatusMovedPermanently), "http status 3xx code to return")
	logFlags := flag.String("log", env("LOG", "json,info"), "log flags, several allowed [debug,info,warn,error,fatal,color,nocolor,json]")
	configPath := flag.String("config", env("CONFIG", ""), "optional JSON file mapping URI prefixes to redirect overrides")
	allowedHosts := flag.String("allowed_hosts", env("ALLOWED_HOSTS", ""), "comma-separated hostnames permitted in Location (required unless -target_host is set)")
	flag.Parse()

	return server.Config{
		TargetProto:    *targetProto,
		TargetHost:     *targetHost,
		TargetPort:     *targetPort,
		TargetPath:     *targetPath,
		BlockQuery:     *blockQuery,
		RedirectStatus: *redirectStatus,
		AllowedHosts:   server.ParseAllowedHosts(*allowedHosts),
	}, *listen, *logFlags, *configPath
}

func defaultListen() string {
	if v := os.Getenv("LISTEN"); v != "" {
		return v
	}
	if v := os.Getenv("PORT"); v != "" {
		if strings.Contains(v, ":") {
			return v
		}
		return ":" + v
	}
	return ":8080"
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func configureLog(flags string) {
	level := slog.LevelInfo
	json := false
	for _, f := range strings.Split(flags, ",") {
		switch strings.TrimSpace(strings.ToLower(f)) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error", "fatal":
			level = slog.LevelError
		case "json":
			json = true
		case "color", "nocolor":
			json = false
		}
	}

	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if json {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))
}
