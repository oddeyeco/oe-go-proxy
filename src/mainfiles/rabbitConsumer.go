package mainfiles

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var conn *amqp.Connection
var connected bool

func connectme() {
	switch connected {
	case false:
		conn, _ = amqp.Dial(to.rabbitAccess)
		connected = true
	}
}
func runConsumer(in bool) {
	connectme()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"oemetrics",                        // name
		true,                               // durable
		false,                              // delete when unused
		false,                              // exclusive
		false,                              // no-wait
		amqp.Table{"x-queue-mode": "lazy"}, // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, _ := ch.Consume(
		q.Name,        // queue
		"oeconsumerg", // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	switch in {
	case true:
		for d := range msgs {
			postData(string(d.Body))
		}
	case false:
		fmt.Println("Lost connection to upstream. Closing AMPQ connection, will try to re-connect to upstream in background .....")
		_ = ch.Cancel("oeconsumerg", true)
		_ = ch.Close()
		_ = conn.Close()
		connected = false
		return

	}

}
