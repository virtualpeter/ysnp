package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config controls how incoming requests are turned into redirect Locations.
type Config struct {
	TargetProto    string
	TargetHost     string
	TargetPort     string
	TargetPath     string
	BlockQuery     bool
	RedirectStatus int
	AllowedHosts   []string
	Routes         map[string]Route
}

// ValidStatus reports whether code is an allowed redirect status.
func ValidStatus(code int) bool {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// ParseAllowedHosts splits a comma-separated hostname list. Empty parts are dropped.
func ParseAllowedHosts(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// Validate reports whether the process-wide host policy is usable.
func (c Config) Validate() error {
	if c.TargetHost != "" {
		if err := checkAllowedHost(c.TargetHost); err != nil {
			return fmt.Errorf("target_host: %w", err)
		}
	}
	for _, h := range c.AllowedHosts {
		if err := checkAllowedHost(h); err != nil {
			return fmt.Errorf("allowed_hosts %q: %w", h, err)
		}
	}
	if c.TargetHost == "" && len(c.AllowedHosts) == 0 {
		return fmt.Errorf("set -target_host or -allowed_hosts")
	}
	return nil
}

func checkAllowedHost(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("empty host")
	}
	if strings.Contains(s, "://") || strings.ContainsAny(s, "/@") {
		return fmt.Errorf("must be a hostname, not a URL")
	}
	return nil
}

// Location builds a fresh redirect URL for r. The destination hostname must be allow-listed.
func (c Config) Location(r *http.Request) (string, error) {
	scheme := "https"
	if c.TargetProto == "http" {
		scheme = "http"
	}

	dest := c.TargetHost
	if dest == "" {
		dest = r.Host
		if dest == "" {
			dest = r.URL.Host
		}
	}
	dest = hostname(dest)
	if dest == "" {
		return "", fmt.Errorf("empty destination host")
	}
	if !c.hostAllowed(dest) {
		return "", fmt.Errorf("host %q is not allowed", dest)
	}

	path := r.URL.Path
	if c.TargetPath != "" {
		path = c.TargetPath
	}
	rawQuery := r.URL.RawQuery
	if c.BlockQuery {
		rawQuery = ""
	}

	u := url.URL{
		Scheme:   scheme,
		Host:     locationHost(dest, c.TargetPort),
		Path:     path,
		RawQuery: rawQuery,
	}
	return u.String(), nil
}

func (c Config) hostAllowed(host string) bool {
	host = hostname(host)
	if c.TargetHost != "" && strings.EqualFold(host, hostname(c.TargetHost)) {
		return true
	}
	for _, allowed := range c.AllowedHosts {
		if strings.EqualFold(host, hostname(strings.TrimSpace(allowed))) {
			return true
		}
	}
	return false
}

func locationHost(name, port string) string {
	if port != "" {
		return net.JoinHostPort(name, port)
	}
	if strings.Contains(name, ":") {
		return "[" + name + "]"
	}
	return name
}

func hostname(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return strings.Trim(host, "[]")
}

// Handler redirects every request according to cfg.
func Handler(cfg Config) http.Handler {
	if cfg.RedirectStatus == 0 {
		cfg.RedirectStatus = http.StatusMovedPermanently
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resolved, _ := cfg.resolve(r.URL.Path)
		loc, err := resolved.Location(r)
		if err != nil {
			http.Error(w, "invalid redirect target", http.StatusBadRequest)
			return
		}
		slog.Debug("redirect", "location", loc)
		http.Redirect(w, r, loc, resolved.RedirectStatus)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Logger logs each request in JSON-friendly structured fields.
func Logger(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		_, log := cfg.resolve(r.URL.Path)
		if log.Skip {
			return
		}
		level := slog.LevelInfo
		if log.Set {
			level = log.Level
		}
		slog.Log(r.Context(), level, "request",
			"remoteAddr", r.RemoteAddr,
			"requestTime", t.Format(time.RFC3339),
			"method", r.Method,
			"requestURL", r.URL.Path,
			"proto", r.Proto,
			"status", sw.status,
			"userAgent", r.UserAgent(),
		)
	})
}
