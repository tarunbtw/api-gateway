package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	routes := []Route{
		{Prefix: "/api/users", Target: "http://localhost:3001"},
		{Prefix: "/api/orders", Target: "http://localhost:3002"},
	}

	proxy := NewProxy(routes)
	limiter := NewRateLimiter(10, 2) // 10 tokens max, refill 2 per second

	// chain: rate limit --> auth --> proxy
	handler := limiter.Middleware(authMiddleware(proxy))

	http.Handle("/", handler)

	// token generator endpoint for testing (no auth required)
	http.HandleFunc("/token", handleToken)

	log.Println("gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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