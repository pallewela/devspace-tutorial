package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	fmt.Fprintf(w, "Hello from DevSpace! Environment: %s\n", env)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/health", healthHandler)

	fmt.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
