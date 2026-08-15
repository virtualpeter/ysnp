package server

import (
	"log/slog"
	"net"
	"net/http"
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

// Location builds the redirect URL for r.
func (c Config) Location(r *http.Request) string {
	u := *r.URL

	if c.TargetProto == "http" {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}

	host := c.TargetHost
	if host == "" {
		host = r.Host
		if host == "" {
			host = r.URL.Host
		}
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}
	if c.TargetPort != "" {
		host = net.JoinHostPort(hostname(host), c.TargetPort)
	}
	u.Host = host

	if c.TargetPath != "" {
		u.Path = c.TargetPath
	}
	if c.BlockQuery {
		u.RawQuery = ""
	}

	return u.String()
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
		loc := resolved.Location(r)
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
