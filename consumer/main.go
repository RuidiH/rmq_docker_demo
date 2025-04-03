package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

// Album represents the album data.
type Album struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// Consumer encapsulates database and RabbitMQ connections.
type Consumer struct {
	db        *sql.DB
	amqpConn  *amqp.Connection
	queueName string
}

// NewConsumer initializes the Consumer with MySQL and RabbitMQ connections.
func NewConsumer(dbDSN, amqpURL, queueName string) (*Consumer, error) {
	// Connect to MySQL.
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("opening MySQL: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging MySQL: %w", err)
	}

	// Connect to RabbitMQ.
	amqpConn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dialing RabbitMQ: %w", err)
	}

	return &Consumer{
		db:        db,
		amqpConn:  amqpConn,
		queueName: queueName,
	}, nil
}

// Close closes the MySQL and RabbitMQ connections.
func (c *Consumer) Close() {
	if c.db != nil {
		c.db.Close()
	}
	if c.amqpConn != nil {
		c.amqpConn.Close()
	}
}

// processAlbum handles the business logic for processing an album.
func (c *Consumer) processAlbum(album Album) error {
	query := "INSERT INTO albums (id, title, artist) VALUES (?, ?, ?)"
	_, err := c.db.Exec(query, album.ID, album.Title, album.Artist)
	return err
}

// Start begins consuming messages from the RabbitMQ queue.
func (c *Consumer) Start() error {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("opening channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declaring queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer tag
		true,  // auto-acknowledge
		false, // exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("registering consumer: %w", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			var album Album
			if err := json.Unmarshal(d.Body, &album); err != nil {
				log.Printf("Error decoding message: %v", err)
				continue
			}
			if err := c.processAlbum(album); err != nil {
				log.Printf("Error inserting album into database: %v", err)
			} else {
				log.Printf("Successfully processed album: %+v", album)
			}
		}
	}()

	fmt.Println("Consumer is waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}

func main() {
	dbDSN := "root:example@tcp(mysql:3306)/albumdb"
	amqpURL := "amqp://guest:guest@rabbitmq:5672/"
	queueName := "album_queue"

	consumer, err := NewConsumer(dbDSN, amqpURL, queueName)
	if err != nil {
		log.Fatalf("Error initializing consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(); err != nil {
		log.Fatalf("Error running consumer: %v", err)
	}
}
