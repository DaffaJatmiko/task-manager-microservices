package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/DaffaJatmiko/listener-service/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// tyr to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
	}
	defer rabbitConn.Close()

	// start listening for messages
	log.Println("Listening for and consuming RabbitMQ messages")


	// create a consumer
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Println(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	// dont continue until rabbitmq is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Printf("backing off %v", backoff)
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}