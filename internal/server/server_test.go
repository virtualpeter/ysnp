package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidStatus(t *testing.T) {
	t.Parallel()
	for _, code := range []int{301, 302, 307, 308} {
		if !ValidStatus(code) {
			t.Errorf("ValidStatus(%d) = false, want true", code)
		}
	}
	if ValidStatus(200) {
		t.Error("ValidStatus(200) = true, want false")
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	if err := (Config{}).Validate(); err == nil {
		t.Fatal("expected error when neither host setting is set")
	}
	if err := (Config{TargetHost: "example.com"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Config{AllowedHosts: []string{"example.com"}}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Config{AllowedHosts: []string{"https://example.com/"}}).Validate(); err == nil {
		t.Fatal("expected error for URL-shaped allow-list entry")
	}
	if err := (Config{TargetHost: "https://example.com/"}).Validate(); err == nil {
		t.Fatal("expected error for URL-shaped target_host")
	}
	if err := (Config{AllowedHosts: []string{"user@example.com"}}).Validate(); err == nil {
		t.Fatal("expected error for userinfo-shaped host")
	}
}

func TestParseAllowedHosts(t *testing.T) {
	t.Parallel()

	if got := ParseAllowedHosts(""); got != nil {
		t.Errorf("empty = %v, want nil", got)
	}
	got := ParseAllowedHosts(" example.com, ,www.example.com ")
	if len(got) != 2 || got[0] != "example.com" || got[1] != "www.example.com" {
		t.Errorf("got %v", got)
	}
}

func TestLocation(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo?q=1", nil)
	req.Host = "example.com"

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default https, keep path and query",
			cfg:  Config{TargetProto: "https", AllowedHosts: []string{"example.com"}},
			want: "https://example.com/foo?q=1",
		},
		{
			name: "override host",
			cfg:  Config{TargetProto: "https", TargetHost: "secure.example.com"},
			want: "https://secure.example.com/foo?q=1",
		},
		{
			name: "override port",
			cfg:  Config{TargetProto: "https", TargetPort: "8443", AllowedHosts: []string{"example.com"}},
			want: "https://example.com:8443/foo?q=1",
		},
		{
			name: "override path",
			cfg:  Config{TargetProto: "https", TargetPath: "/landing", AllowedHosts: []string{"example.com"}},
			want: "https://example.com/landing?q=1",
		},
		{
			name: "block query",
			cfg:  Config{TargetProto: "https", BlockQuery: true, AllowedHosts: []string{"example.com"}},
			want: "https://example.com/foo",
		},
		{
			name: "http proto",
			cfg:  Config{TargetProto: "http", AllowedHosts: []string{"example.com"}},
			want: "http://example.com/foo?q=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.cfg.Location(req)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Location() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocationUsesRequestHostWhenURLHostEmpty(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Host = "example.com:8080"

	got, err := Config{TargetProto: "https", AllowedHosts: []string{"example.com"}}.Location(req)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/foo"
	if got != want {
		t.Errorf("Location() = %q, want %q", got, want)
	}
}

func TestLocationRejectsUnknownAndCybersquatHosts(t *testing.T) {
	t.Parallel()

	cfg := Config{TargetProto: "https", AllowedHosts: []string{"example.com"}}

	req := httptest.NewRequest(http.MethodGet, "http://evil.example/foo", nil)
	req.Host = "evil.example"
	if _, err := cfg.Location(req); err == nil {
		t.Fatal("expected error for unknown host")
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com.evil.org/foo", nil)
	req.Host = "example.com.evil.org"
	if _, err := cfg.Location(req); err == nil {
		t.Fatal("expected error for cybersquat host")
	}
}

func TestLocationTargetHostIgnoresSpoofedHost(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://evil.example/foo", nil)
	req.Host = "evil.example"

	got, err := Config{TargetProto: "https", TargetHost: "example.com"}.Location(req)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/foo" {
		t.Errorf("Location() = %q, want https://example.com/foo", got)
	}
}

func TestLocationHasNoUserinfo(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Host = "evil.com@example.com"

	cfg := Config{TargetProto: "https", AllowedHosts: []string{"example.com"}}
	if _, err := cfg.Location(req); err == nil {
		t.Fatal("expected error for userinfo-shaped Host")
	}

	req.Host = "example.com"
	got, err := Config{TargetProto: "https", TargetHost: "example.com"}.Location(req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "@") {
		t.Errorf("Location contains userinfo: %q", got)
	}
}

func TestHandlerRedirects(t *testing.T) {
	t.Parallel()

	h := Handler(Config{
		TargetProto:    "https",
		RedirectStatus: http.StatusFound,
		AllowedHosts:   []string{"example.com"},
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/bar", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/bar" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com/bar")
	}
}

func TestHandlerRejectsDisallowedHost(t *testing.T) {
	t.Parallel()

	h := Handler(Config{TargetProto: "https", AllowedHosts: []string{"example.com"}})
	req := httptest.NewRequest(http.MethodGet, "http://evil.example/bar", nil)
	req.Host = "evil.example"
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, want empty", loc)
	}
}
