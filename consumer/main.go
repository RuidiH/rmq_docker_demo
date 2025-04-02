package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

type Album struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

func main() {
	// Connect to RabbitMQ using the service name 'rabbitmq'
	amqpConn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer amqpConn.Close()

	ch, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"album_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	// Connect to MySQL using the service name 'mysql'
	db, err := sql.Open("mysql", "root:example@tcp(mysql:3306)/albumdb")
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %s", err)
	}
	defer db.Close()

	// Ensure the database connection is alive
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var album Album
			if err := json.Unmarshal(d.Body, &album); err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}
			// Insert the album into the database
			query := "INSERT INTO albums (id, title, artist) VALUES (?, ?, ?)"
			_, err := db.Exec(query, album.ID, album.Title, album.Artist)
			if err != nil {
				log.Printf("Failed to insert album: %s", err)
			} else {
				log.Printf("Inserted album: %+v", album)
			}
		}
	}()

	fmt.Println("Consumer is waiting for messages. To exit press CTRL+C")
	<-forever
}
