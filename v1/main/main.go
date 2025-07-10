package main

import (
	"dot/v1/objects"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/objects/", objects.Handler)
	// addr := os.Getenv("LISTEN_ADDRESS")
	// err := http.ListenAndServe(addr, nil)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}