package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

type Album struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

var amqpConn *amqp.Connection

func main() {
	// Connect to RabbitMQ using the service name 'rabbitmq'
	var err error
	amqpConn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer amqpConn.Close()

	http.HandleFunc("/album", albumHandler)
	log.Println("Publisher running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func albumHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Return a dummy album info as JSON
		album := Album{ID: 1, Title: "Album Title", Artist: "Artist Name"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(album)
	case "POST":
		var album Album
		if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		// Publish album info to RabbitMQ
		if err := publishAlbum(album); err != nil {
			http.Error(w, "Failed to publish album", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Album published")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func publishAlbum(album Album) error {
	ch, err := amqpConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare a queue named "album_queue"
	q, err := ch.QueueDeclare(
		"album_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(album)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	return err
}
