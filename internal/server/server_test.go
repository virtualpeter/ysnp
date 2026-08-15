package server

import (
	"net/http"
	"net/http/httptest"
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
			cfg:  Config{TargetProto: "https"},
			want: "https://example.com/foo?q=1",
		},
		{
			name: "override host",
			cfg:  Config{TargetProto: "https", TargetHost: "secure.example.com"},
			want: "https://secure.example.com/foo?q=1",
		},
		{
			name: "override port",
			cfg:  Config{TargetProto: "https", TargetPort: "8443"},
			want: "https://example.com:8443/foo?q=1",
		},
		{
			name: "override path",
			cfg:  Config{TargetProto: "https", TargetPath: "/landing"},
			want: "https://example.com/landing?q=1",
		},
		{
			name: "block query",
			cfg:  Config{TargetProto: "https", BlockQuery: true},
			want: "https://example.com/foo",
		},
		{
			name: "http proto",
			cfg:  Config{TargetProto: "http"},
			want: "http://example.com/foo?q=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.cfg.Location(req)
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

	got := Config{TargetProto: "https"}.Location(req)
	want := "https://example.com/foo"
	if got != want {
		t.Errorf("Location() = %q, want %q", got, want)
	}
}

func TestHandlerRedirects(t *testing.T) {
	t.Parallel()

	h := Handler(Config{TargetProto: "https", RedirectStatus: http.StatusFound})
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
