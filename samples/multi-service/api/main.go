package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Hello from API!"}`)
}

func main() {
	http.HandleFunc("/api", handler)
	fmt.Println("API listening on :8080")
	http.ListenAndServe(":8080", nil)
}
