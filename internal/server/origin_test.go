package server

import (
	"net/http"
	"testing"
)

func TestParseOrigins(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"*", []string{"*"}},
		{"", []string{"*"}},
		{" http://localhost:8080 , http://127.0.0.1:5173 ", []string{
			"http://localhost:8080",
			"http://127.0.0.1:5173",
		}},
	}
	for _, tc := range cases {
		got := ParseOrigins(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("ParseOrigins(%q) len=%d want %d", tc.in, len(got), len(tc.want))
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("ParseOrigins(%q)[%d]=%q want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

func TestCheckOriginFunc(t *testing.T) {
	allowLocal := CheckOriginFunc([]string{"http://localhost:8080", "http://localhost:5173"})
	allowAll := CheckOriginFunc([]string{"*"})

	req := func(origin string) *http.Request {
		r, _ := http.NewRequest(http.MethodGet, "http://example/ws", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return r
	}

	if !allowLocal(req("")) {
		t.Fatal("empty Origin should be allowed")
	}
	if !allowLocal(req("http://localhost:8080")) {
		t.Fatal("listed Origin should be allowed")
	}
	if allowLocal(req("http://evil.example")) {
		t.Fatal("unlisted Origin should be rejected")
	}
	if !allowAll(req("http://evil.example")) {
		t.Fatal("* should allow any Origin")
	}
}
