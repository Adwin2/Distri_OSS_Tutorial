package heartbeat

import (
	"dot/v2/rabbitmq"
	"log"
	"os"
	"time"
)

func StartHeartbeat() {
	server := os.Getenv("RABBITMQ_SERVER")

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Heartbeat panic recovered: %v", r)
					time.Sleep(5 * time.Second) // 等待后重试
				}
			}()

			q := rabbitmq.New(server)
			defer q.Close()

			for {
				address := os.Getenv("LISTEN_ADDRESS")
				q.Publish("apiServers", address)
				time.Sleep(5 * time.Second)
			}
		}()
	}
}