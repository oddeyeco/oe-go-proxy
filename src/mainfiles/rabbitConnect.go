package mainfiles

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

type rqueue struct {
	url  string
	name string

	errorChannel chan *amqp.Error
	connection   *amqp.Connection
	channel      *amqp.Channel
	closed       bool

	consumers []messageConsumer
}

type messageConsumer func(string)

func NewQueue(url string, qName string) *rqueue {
	q := new(rqueue)
	q.url = url
	q.name = qName
	q.consumers = make([]messageConsumer, 0)

	q.connect()
	go q.reconnector()

	return q
}

func (q *rqueue) Send(message []byte) {
	err := q.channel.Publish(
		"",     // exchange
		q.name, // routing key
		false,  // mandatory
		false,  // immediate

		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		})
	logError("Sending message to rqueue failed", err)
}

func (q *rqueue) Consume(consumer messageConsumer) {
	deliveries, err := q.registerQueueConsumer()
	log.Println("Consumer registered! Processing messages...")
	q.executeMessageConsumer(err, consumer, deliveries, false)
}

func (q *rqueue) Close() {
	//log.Println("Closing connection")
	q.closed = true
	q.channel.Close()
	q.connection.Close()
}

func (q *rqueue) reconnector() {
	for {
		err := <-q.errorChannel
		if !q.closed {
			logError("Reconnecting after connection closed", err)

			q.connect()
			q.recoverConsumers()
		}
	}
}

func (q *rqueue) connect() {
	for {
		//log.Printf("Connecting to rabbitmq on %s\n", q.url)
		conn, err := amqp.Dial(q.url)
		if err == nil {
			q.connection = conn
			q.errorChannel = make(chan *amqp.Error)
			q.connection.NotifyClose(q.errorChannel)

			//log.Println("Connection established!")

			q.openChannel()
			q.declareQueue()

			return
		}

		logError("Connection to rabbitmq failed. Retrying in 1 sec... ", err)
		time.Sleep(1000 * time.Millisecond)
	}
}

func (q *rqueue) declareQueue() {
	_, err := q.channel.QueueDeclare(
		q.name,                             // name
		true,                               // durable
		false,                              // delete when unused
		false,                              // exclusive
		false,                              // no-wait
		amqp.Table{"x-queue-mode": "lazy"}, // arguments
	)
	logError("Queue declaration failed", err)
}

func (q *rqueue) openChannel() {
	channel, err := q.connection.Channel()
	logError("Opening channel failed", err)
	q.channel = channel
}

func (q *rqueue) registerQueueConsumer() (<-chan amqp.Delivery, error) {
	msgs, err := q.channel.Consume(
		q.name, // rqueue
		"",     // messageConsumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	logError("Consuming messages from rqueue failed", err)
	return msgs, err
}
func (q *rqueue) executeMessageConsumer(err error, consumer messageConsumer, deliveries <-chan amqp.Delivery, isRecovery bool) {
	if err == nil {
		if !isRecovery {
			q.consumers = append(q.consumers, consumer)
		}
		go func() {
			for delivery := range deliveries {
				consumer(string(delivery.Body[:]))
			}
		}()
	}
}

func (q *rqueue) recoverConsumers() {
	for i := range q.consumers {
		var consumer = q.consumers[i]

		log.Println("Recovering consumer...")
		msgs, err := q.registerQueueConsumer()
		log.Println("Consumer recovered! Continuing message processing...")
		q.executeMessageConsumer(err, consumer, msgs, true)
	}
}

func logError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s", message, err)
	}
}
