package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

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

	handler := limiter.Middleware(authMiddleware(proxy))

	http.Handle("/", handler)
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