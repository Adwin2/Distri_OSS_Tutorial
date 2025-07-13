package main

import (
	"dot/v2/dataserver/heartbeat"
	"dot/v2/dataserver/locate"
	"dot/v2/objects"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("/home/raymond/桌面/expr/Distri_OSS_Tutorial/v2/.env"); err != nil {
		log.Print(err)
	}
	// 心跳
	go heartbeat.StartHeartbeat()
	// 定位对象
	go locate.StartLocate()

	http.HandleFunc("/objects/", objects.Handler)
	address := os.Getenv("LISTEN_ADDRESS")
	http.ListenAndServe(address, nil)
}