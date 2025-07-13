package locate

import (
	"dot/v2/rabbitmq"
	"log"
	"os"
	"strconv"
	"time"
)

func Locate(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func StartLocate() {
	server := os.Getenv("RABBITMQ_SERVER")

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Locate service panic recovered: %v", r)
					time.Sleep(5 * time.Second) // 等待后重试
				}
			}()

			q := rabbitmq.New(server)
			defer q.Close()
			q.Bind("dataServers")

			for msg := range q.Consume() {
				object, err := strconv.Unquote(string(msg.Body))
				if err != nil {
					log.Printf("Failed to unquote message: %v", err)
					continue
				}
				root := os.Getenv("STORAGE_ROOT")
				name := root + "/objects/" + object
				if Locate(name) {
					q.Send(msg.ReplyTo, name)
				}
			}
		}()
	}
}