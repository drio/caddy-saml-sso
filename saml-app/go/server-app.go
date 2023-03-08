package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cn := r.Header.Get("cn")
		fmt.Fprintf(w, fmt.Sprintf("You are authenticated: [%s]\n", cn))
	})

	port := 8182
	fmt.Printf("Listening on :%d...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("I cannot start server: %s", err)
	}
}
