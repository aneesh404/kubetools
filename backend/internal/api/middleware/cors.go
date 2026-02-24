package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

func CORS(origins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	isAllowed := func(origin string) bool {
		if origin == "" {
			return false
		}
		if _, ok := allowed["*"]; ok {
			return true
		}
		if _, ok := allowed[origin]; ok {
			return true
		}

		parsed, err := url.Parse(origin)
		if err != nil {
			return false
		}

		host := strings.ToLower(parsed.Hostname())
		switch host {
		case "localhost", "127.0.0.1", "0.0.0.0", "::1":
			return true
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return false
		}

		if ip.IsLoopback() || ip.IsPrivate() {
			return true
		}

		return false
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
