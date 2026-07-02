package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// startUsersService starts a minimal users service on 127.0.0.1:3001.
// It runs as a goroutine inside the gateway process so the whole application
// ships as a single binary. In a production setup this would be a separate
// deployable service.
func startUsersService() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"service": "users",
			"path":    r.URL.Path,
			"method":  r.Method,
			"message": "hello from users service",
		})
	})
	log.Println("users service listening on 127.0.0.1:3001 (in-process)")
	if err := http.ListenAndServe("127.0.0.1:3001", mux); err != nil {
		log.Fatalf("users service error: %v", err)
	}
}

// startOrdersService starts a minimal orders service on 127.0.0.1:3002.
// Same consolidation rationale as startUsersService.
func startOrdersService() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"service": "orders",
			"path":    r.URL.Path,
			"method":  r.Method,
			"message": "hello from orders service",
		})
	})
	log.Println("orders service listening on 127.0.0.1:3002 (in-process)")
	if err := http.ListenAndServe("127.0.0.1:3002", mux); err != nil {
		log.Fatalf("orders service error: %v", err)
	}
}
