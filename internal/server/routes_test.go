package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPrefixMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key, path string
		want      bool
	}{
		{key: "/", path: "/", want: true},
		{key: "/", path: "/anything", want: true},
		{key: "/api", path: "/api", want: true},
		{key: "/api", path: "/api/", want: true},
		{key: "/api", path: "/api/v1", want: true},
		{key: "/api", path: "/apiv2", want: false},
		{key: "/api/", path: "/api", want: false},
		{key: "/api/", path: "/api/", want: true},
		{key: "/api/", path: "/api/v1", want: true},
		{key: "/old", path: "/older", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.key+" "+tt.path, func(t *testing.T) {
			t.Parallel()
			if got := prefixMatch(tt.key, tt.path); got != tt.want {
				t.Errorf("prefixMatch(%q, %q) = %v, want %v", tt.key, tt.path, got, tt.want)
			}
		})
	}
}

func TestLongestPrefixWins(t *testing.T) {
	t.Parallel()

	routes := map[string]Route{
		"/api":    {TargetPath: strPtr("/short")},
		"/api/v1": {TargetPath: strPtr("/long")},
		"/":       {TargetPath: strPtr("/root")},
	}

	got, ok := longestPrefix(routes, "/api/v1/users")
	if !ok {
		t.Fatal("expected a match")
	}
	if got.TargetPath == nil || *got.TargetPath != "/long" {
		t.Errorf("got %#v, want /long", got.TargetPath)
	}

	got, ok = longestPrefix(routes, "/other")
	if !ok {
		t.Fatal("expected catch-all")
	}
	if got.TargetPath == nil || *got.TargetPath != "/root" {
		t.Errorf("got %#v, want /root", got.TargetPath)
	}
}

func TestResolveInheritsAndOverrides(t *testing.T) {
	t.Parallel()

	base := Config{
		TargetProto:    "https",
		TargetPort:     "443",
		TargetPath:     "",
		BlockQuery:     false,
		RedirectStatus: http.StatusMovedPermanently,
		Routes: map[string]Route{
			"/old": {
				TargetPath: strPtr("/new"),
				TargetPort: strPtr("8443"),
				BlockQuery: boolPtr(true),
				Status:     intPtr(http.StatusFound),
				Log:        strPtr("debug"),
			},
			"/quiet": {
				Log: strPtr("off"),
			},
			"/keep": {},
		},
	}

	over, log := base.resolve("/old/page")
	if over.TargetPath != "/new" || over.TargetPort != "8443" || !over.BlockQuery || over.RedirectStatus != http.StatusFound {
		t.Errorf("override = %+v", over)
	}
	if over.TargetProto != "https" {
		t.Errorf("TargetProto = %q, want inherited https", over.TargetProto)
	}
	if !log.Set || log.Level != slog.LevelDebug || log.Skip {
		t.Errorf("log = %+v, want debug", log)
	}

	quiet, log := base.resolve("/quiet")
	if quiet.TargetPort != "443" || quiet.BlockQuery || quiet.RedirectStatus != http.StatusMovedPermanently {
		t.Errorf("quiet inherit = %+v", quiet)
	}
	if !log.Skip {
		t.Errorf("log = %+v, want skip", log)
	}

	keep, log := base.resolve("/keep")
	if keep.TargetPort != "443" || keep.TargetPath != "" {
		t.Errorf("keep inherit = %+v", keep)
	}
	if log.Skip || log.Set {
		t.Errorf("log = %+v, want default", log)
	}

	miss, log := base.resolve("/unlisted")
	if miss.TargetPort != "443" || miss.RedirectStatus != http.StatusMovedPermanently {
		t.Errorf("miss = %+v", miss)
	}
	if log.Skip || log.Set {
		t.Errorf("miss log = %+v, want default", log)
	}
}

func TestParseLogOverride(t *testing.T) {
	t.Parallel()

	if got := parseLogOverride(nil); got.Skip || got.Set {
		t.Errorf("nil = %+v", got)
	}
	if got := parseLogOverride(strPtr("off")); !got.Skip {
		t.Errorf("off = %+v", got)
	}
	if got := parseLogOverride(strPtr("warn,json")); !got.Set || got.Level != slog.LevelWarn || got.Skip {
		t.Errorf("warn,json = %+v", got)
	}
	if got := parseLogOverride(strPtr("debug,off")); !got.Skip {
		t.Errorf("debug,off = %+v, want skip", got)
	}
}

func TestLoadRoutesEmptyPath(t *testing.T) {
	t.Parallel()

	routes, err := LoadRoutes("")
	if err != nil {
		t.Fatal(err)
	}
	if routes != nil {
		t.Errorf("got %v, want nil", routes)
	}
}

func TestLoadRoutesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "routes.json")
	if err := os.WriteFile(path, []byte(`{
		"/old": {"target_path": "/new", "status": 302, "blockquery": true}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	routes, err := LoadRoutes(path)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := routes["/old"]
	if !ok {
		t.Fatal("missing /old")
	}
	if r.TargetPath == nil || *r.TargetPath != "/new" {
		t.Errorf("target_path = %v", r.TargetPath)
	}
	if r.Status == nil || *r.Status != 302 {
		t.Errorf("status = %v", r.Status)
	}
	if r.BlockQuery == nil || !*r.BlockQuery {
		t.Errorf("blockquery = %v", r.BlockQuery)
	}
}

func TestLoadRoutesErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte(`{`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadRoutes(path); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "emptykey.json")
		if err := os.WriteFile(path, []byte(`{"": {"status": 301}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadRoutes(path); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("bad status", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "status.json")
		if err := os.WriteFile(path, []byte(`{"/x": {"status": 200}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadRoutes(path); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		if _, err := LoadRoutes(filepath.Join(dir, "nope.json")); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestHandlerUsesRoute(t *testing.T) {
	t.Parallel()

	h := Handler(Config{
		TargetProto:    "https",
		RedirectStatus: http.StatusMovedPermanently,
		AllowedHosts:   []string{"example.com"},
		Routes: map[string]Route{
			"/old": {
				TargetPath: strPtr("/new"),
				Status:     intPtr(http.StatusFound),
				BlockQuery: boolPtr(true),
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/old?q=1", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/new" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com/new")
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com/other?q=1", nil)
	req.Host = "example.com"
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("unmatched status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/other?q=1" {
		t.Errorf("unmatched Location = %q", loc)
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
