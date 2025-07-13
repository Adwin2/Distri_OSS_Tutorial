package rabbitmq

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ 结构体
type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	Name     string
	exchange string
}

// New 创建一个 RabbitMQ 对象
func New(s string) *RabbitMQ {
	mq := &RabbitMQ{}
	err := mq.connect(s)
	if err != nil {
		panic(err)
	}
	return mq
}

// connect 建立连接
func (q *RabbitMQ) connect(s string) error {
	// 创建连接
	conn, err := amqp.Dial(s)
	if err != nil {
		return err
	}

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	// 创建队列
	queue, err := ch.QueueDeclare(
		"",      // name
		false,   // durable
		true,    // autoDelete
		false,   // exclusive
		false,   // noWait
		nil,     // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	q.conn = conn
	q.channel = ch
	q.Name = queue.Name
	return nil
}

// isConnected 检查连接是否有效
func (q *RabbitMQ) isConnected() bool {
	return q.conn != nil && !q.conn.IsClosed() && q.channel != nil
}

// Bind 绑定交换机
func (q *RabbitMQ) Bind (exchange string) {
	// 尝试绑定队列到交换机
	// 注意：这里假设交换机已经存在，如果不存在会返回错误
	err := q.channel.QueueBind(
		q.Name, "", exchange, false, nil,
	)
	if err != nil {
		// 如果交换机不存在，尝试创建它（仅用于开发环境）
		if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == 404 {
			// 声明交换机
			declareErr := q.channel.ExchangeDeclare(
				exchange, // name
				"fanout", // type
				true,     // durable
				false,    // auto-deleted
				false,    // internal
				false,    // no-wait
				nil,      // arguments
			)
			if declareErr != nil {
				panic(declareErr)
			}

			// 重新尝试绑定
			err = q.channel.QueueBind(
				q.Name, "", exchange, false, nil,
			)
		}

		if err != nil {
			panic(err)
		}
	}
	q.exchange = exchange
}

func (q *RabbitMQ) Send(queue string, body any) {
	s, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	// match the binding routing key as ""
	err = q.channel.Publish(
		"",queue, false, false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body: []byte(s),
		},
	)
	if err != nil {
		panic(err)
	}
}

func (q *RabbitMQ) Publish(exchange string, body any) {
	s, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	// 尝试发布消息
	err = q.channel.Publish(
		exchange, "", false, false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body: []byte(s),
		},
	)

	// 如果交换机不存在，尝试创建它（仅用于开发环境）
	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == 404 {
			// 声明交换机
			declareErr := q.channel.ExchangeDeclare(
				exchange, // name
				"fanout", // type
				true,     // durable
				false,    // auto-deleted
				false,    // internal
				false,    // no-wait
				nil,      // arguments
			)
			if declareErr != nil {
				panic(declareErr)
			}

			// 重新尝试发布
			err = q.channel.Publish(
				exchange, "", false, false,
				amqp.Publishing{
					ReplyTo: q.Name,
					Body: []byte(s),
				},
			)
		}

		if err != nil {
			panic(err)
		}
	}
}

func (q * RabbitMQ) Consume() <-chan amqp.Delivery {
	c, err := q.channel.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		panic(err)
	}
	return c
}

// Close 关闭通道和连接
func (q *RabbitMQ) Close() {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		q.conn.Close()
	}
}