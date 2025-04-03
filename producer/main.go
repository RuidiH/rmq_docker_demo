package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/streadway/amqp"
)

// Album represents the album data.
type Album struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// Publisher encapsulates the AMQP connection.
type Producer struct {
	amqpConn *amqp.Connection
}

// NewPublisher creates a new Publisher by connecting to RabbitMQ.
func NewProducer(amqpURL string) (*Producer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dialing RabbitMQ: %w", err)
	}
	return &Producer{amqpConn: conn}, nil
}

// Close closes the AMQP connection.
func (p *Producer) Close() error {
	return p.amqpConn.Close()
}

// PublishAlbum publishes album data to the RabbitMQ queue.
func (p *Producer) PublishAlbum(album Album) error {
	ch, err := p.amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("opening channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"album_queue", // queue name
		true,          // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declaring queue: %w", err)
	}

	body, err := json.Marshal(album)
	if err != nil {
		return fmt.Errorf("marshaling album: %w", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	return nil
}

// albumHandler returns an HTTP handler that handles GET and POST requests.
func albumHandler(p *Producer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// Require an 'id' parameter
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "Missing album id", http.StatusBadRequest)
				return
			}
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid album id", http.StatusBadRequest)
				return
			}
			// In a real scenario, retrieve the album from a datastore.
			album := Album{
				ID:     id,
				Title:  fmt.Sprintf("Album Title %d", id),
				Artist: "Artist Name",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(album)
		case "POST":
			var album Album
			if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			// Basic validation.
			if album.ID == 0 || album.Title == "" || album.Artist == "" {
				http.Error(w, "Missing album fields", http.StatusBadRequest)
				return
			}
			// Publish album to RabbitMQ.
			if err := p.PublishAlbum(album); err != nil {
				http.Error(w, "Failed to publish album", http.StatusInternalServerError)
				log.Printf("Publish error: %v", err)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, "Album published")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	amqpURL := "amqp://guest:guest@rabbitmq:5672/"
	producer, err := NewProducer(amqpURL)
	if err != nil {
		log.Fatalf("Failed to initialize publisher: %v", err)
	}
	defer producer.Close()

	http.HandleFunc("/album", albumHandler(producer))
	log.Println("Publisher running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
