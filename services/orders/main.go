package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"service": "orders",
			"path":    r.URL.Path,
			"method":  r.Method,
			"message": "hello from orders service",
		})
	})

	log.Println("orders service listening on :3002")
	log.Fatal(http.ListenAndServe(":3002", nil))
}