package main

import (
	"fmt"
	"io"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Frontend handler called")
	// Call the API
	resp, err := http.Get("http://api:9090/api")
	if err != nil {
		fmt.Fprintf(w, "Error calling API: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(w, "Frontend calling API: %s", body)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Frontend listening on :3000")
	http.ListenAndServe(":3000", nil)
}
