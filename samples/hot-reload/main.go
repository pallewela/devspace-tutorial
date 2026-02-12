package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handler called @ %s\n", time.Now())
	fmt.Fprintf(w, "Hello hello!\n")
}

func main() {
	fmt.Println("Started server on :9090")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":9090", nil)
}
