package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Route is an optional per-prefix overlay. Nil fields inherit the process-wide Config.
type Route struct {
	TargetPath *string `json:"target_path"`
	TargetPort *string `json:"target_port"`
	BlockQuery *bool   `json:"blockquery"`
	Status     *int    `json:"status"`
	Log        *string `json:"log"`
}

type logOverride struct {
	Skip  bool
	Set   bool
	Level slog.Level
}

// LoadRoutes reads a JSON object of path prefix → Route. An empty path is a no-op.
func LoadRoutes(path string) (map[string]Route, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var routes map[string]Route
	if err := json.Unmarshal(data, &routes); err != nil {
		return nil, err
	}
	if routes == nil {
		routes = map[string]Route{}
	}
	for key, route := range routes {
		if key == "" {
			return nil, fmt.Errorf("empty route key")
		}
		if route.Status != nil && !ValidStatus(*route.Status) {
			return nil, fmt.Errorf("route %q: status must be one of 301, 302, 307, 308", key)
		}
	}
	return routes, nil
}

func (c Config) resolve(path string) (Config, logOverride) {
	route, ok := longestPrefix(c.Routes, path)
	if !ok {
		return c, logOverride{}
	}
	out := c
	if route.TargetPath != nil {
		out.TargetPath = *route.TargetPath
	}
	if route.TargetPort != nil {
		out.TargetPort = *route.TargetPort
	}
	if route.BlockQuery != nil {
		out.BlockQuery = *route.BlockQuery
	}
	if route.Status != nil {
		out.RedirectStatus = *route.Status
	}
	return out, parseLogOverride(route.Log)
}

func longestPrefix(routes map[string]Route, path string) (Route, bool) {
	var (
		best   string
		found  bool
		chosen Route
	)
	for key, route := range routes {
		if !prefixMatch(key, path) {
			continue
		}
		if !found || len(key) > len(best) {
			best = key
			chosen = route
			found = true
		}
	}
	return chosen, found
}

// prefixMatch reports whether key is a path-boundary prefix of path.
// /api matches /api and /api/v1, but not /apiv2. / is a catch-all.
func prefixMatch(key, path string) bool {
	if key == "/" {
		return true
	}
	if path == key {
		return true
	}
	if !strings.HasPrefix(path, key) {
		return false
	}
	if strings.HasSuffix(key, "/") {
		return true
	}
	return len(path) > len(key) && path[len(key)] == '/'
}

func parseLogOverride(log *string) logOverride {
	if log == nil {
		return logOverride{}
	}
	var o logOverride
	for _, f := range strings.Split(*log, ",") {
		switch strings.TrimSpace(strings.ToLower(f)) {
		case "off":
			o.Skip = true
		case "debug":
			o.Set = true
			o.Level = slog.LevelDebug
		case "info":
			o.Set = true
			o.Level = slog.LevelInfo
		case "warn":
			o.Set = true
			o.Level = slog.LevelWarn
		case "error", "fatal":
			o.Set = true
			o.Level = slog.LevelError
		}
	}
	return o
}
