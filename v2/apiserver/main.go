package apiserver

import (
	"dot/v2/apiserver/heartbeat"
	"dot/v2/apiserver/locate"
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
	go heartbeat.ListenHeartbeat()
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	address := os.Getenv("LISTEN_ADDRESS")
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Println(err)
	}
}