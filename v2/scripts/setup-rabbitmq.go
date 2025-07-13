package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("Warning: .env file not found")
	}

	server := os.Getenv("RABBITMQ_SERVER")
	if server == "" {
		server = "amqp://localhost:5672/"
	}

	// 连接到RabbitMQ
	conn, err := amqp.Dial(server)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 创建所需的交换机
	exchanges := []string{"apiServers", "dataServers"}
	
	for _, exchange := range exchanges {
		err = ch.ExchangeDeclare(
			exchange, // name
			"fanout", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare exchange %s: %v", exchange, err)
		}
		log.Printf("Exchange '%s' declared successfully", exchange)
	}

	log.Println("RabbitMQ setup completed successfully!")
}
