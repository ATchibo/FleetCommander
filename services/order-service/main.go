package main

import (
	"encoding/json"
	"log"
	"time"
	"strings" 
	
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Order structure
type Order struct {
	ID          string `json:"id"`
	Customer    string `json:"customer"`
	Destination string `json:"destination"`
	Price       int    `json:"price"`
}

func main() {
	// 1. Connect to RabbitMQ
	// We retry connection because RabbitMQ might take a few seconds to start
	var conn *amqp.Connection
	var err error
	
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err == nil {
			break
		}
		log.Println("Waiting for RabbitMQ...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// 2. Declare the Queue
	q, err := ch.QueueDeclare(
		"orders_queue", // name
		true,           // durable (save to disk)
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Setup Web Server (Fiber)
	app := fiber.New()

	// THE SHARED SECRET (Must match Auth Service)
    var jwtSecret = []byte("super-secret-fleet-key")

	// --- 🔒 SECURITY MIDDLEWARE (JWT) ---
	app.Use(func(c *fiber.Ctx) error {
        // 1. Get Token from Header (Authorization: Bearer <token>)
		authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{"error": "Missing Token"})
        }

        // Remove "Bearer " prefix
        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

        // 2. Parse & Verify Token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            return c.Status(401).JSON(fiber.Map{"error": "Invalid or Expired Token"})
        }

        // Token is valid! Proceed.
		return c.Next()
	})
	// ------------------------------------

	app.Post("/orders", func(c *fiber.Ctx) error {
		// Parse JSON body
		order := new(Order)
		if err := c.BodyParser(order); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Convert to Bytes
		body, _ := json.Marshal(order)

		// 4. Publish to RabbitMQ
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		
		if err != nil {
			return c.Status(500).SendString("Failed to publish order")
		}

		log.Printf("📦 Order Received & Queued: %s to %s", order.Customer, order.Destination)
		return c.JSON(fiber.Map{"status": "queued", "message": "Order sent to driver dispatch"})
	})

	log.Println("🚀 Order Service running on http://localhost:3000")
	// Note: We use port 3001 because 3000 is taken by the WebSocket service
	app.Listen(":3001") 
}