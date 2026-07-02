package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	// Start backend services in-process on loopback addresses.
	// These are reachable only within the process; no external port is exposed for them.
	go startUsersService()
	go startOrdersService()

	usersService := os.Getenv("USERS_SERVICE")
	if usersService == "" {
		usersService = "http://localhost:3001"
	}

	ordersService := os.Getenv("ORDERS_SERVICE")
	if ordersService == "" {
		ordersService = "http://localhost:3002"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		jwtSecret = []byte(secret)
	}

	routes := []Route{
		{Prefix: "/api/users", Target: usersService},
		{Prefix: "/api/orders", Target: ordersService},
	}

	proxy := NewProxy(routes)
	limiter := NewRateLimiter(10, 2)

	// Rate-limited, auth-checked handler for all /api/* traffic.
	apiHandler := limiter.Middleware(authMiddleware(proxy))

	// Use an explicit mux (not the DefaultServeMux) so it is completely isolated
	// from the internal service muxes started above.
	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler)
	mux.HandleFunc("/token", handleToken)
	mux.HandleFunc("/health", handleHealth(routes, limiter))
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// handleToken issues a JWT for the demo user.
// DEMO ONLY: unauthenticated token issuance, replace with real auth (password/OAuth) before production use.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token, err := generateToken("test-user")
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"token":"` + token + `"}`))
}

// handleHealth returns gateway liveness info including configured routes and rate-limit params.
// Render (and uptime monitors) use GET /health to verify the service is alive.
func handleHealth(routes []Route, limiter *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		routePaths := make([]string, len(routes))
		for i, route := range routes {
			routePaths[i] = route.Prefix
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"routes": routePaths,
			"rateLimit": map[string]any{
				"maxTokens":       limiter.max,
				"refillPerSecond": limiter.refill,
			},
		})
	}
}
