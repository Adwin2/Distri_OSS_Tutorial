package main

import (
	"dot/v1/objects"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("/home/raymond/桌面/expr/Distri_OSS_Tutorial/v1/.env"); err != nil {
		log.Print(err)
	}
	http.HandleFunc("/objects/", objects.Handler)
	addr := os.Getenv("LISTEN_ADDRESS")
	log.Println("Listening on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}