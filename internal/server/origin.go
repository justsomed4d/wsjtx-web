package server

import (
	"net/http"
	"strings"
)

// ParseOrigins splits a comma-separated origin allowlist.
// A single "*" means allow any Origin.
func ParseOrigins(spec string) []string {
	spec = strings.TrimSpace(spec)
	if spec == "" || spec == "*" {
		return []string{"*"}
	}
	parts := strings.Split(spec, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

// CheckOriginFunc returns a websocket CheckOrigin callback for the allowlist.
// Empty Origin headers are allowed (non-browser clients). "*" allows all.
func CheckOriginFunc(allowed []string) func(*http.Request) bool {
	allowAll := false
	set := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		if o == "*" {
			allowAll = true
			break
		}
		set[o] = struct{}{}
	}
	return func(r *http.Request) bool {
		if allowAll {
			return true
		}
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := set[origin]
		return ok
	}
}
