package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Secret key used to sign tokens (In prod, keep this in ENV variables!)
var jwtSecret = []byte("super-secret-fleet-key")

// Mock User Database
var users = map[string]string{
	"admin": "password123",
	"driver": "securepass",
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	app := fiber.New()

	// --- ENABLE CORS ---
    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:5173", // Allow your Vue App
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    }))
    // -------------------

	app.Post("/login", func(c *fiber.Ctx) error {
		// 1. Parse Input
		req := new(LoginRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).SendString("Bad Request")
		}

		// 2. Validate User
		pass, exists := users[req.Username]
		if !exists || pass != req.Password {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		// 3. Create Token (The "Passport")
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["user"] = req.Username
		claims["role"] = "admin" // You could change this based on user
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Valid for 3 days

		// 4. Sign the token
		t, err := token.SignedString(jwtSecret)
		if err != nil {
			return c.SendStatus(500)
		}

		return c.JSON(fiber.Map{"token": t})
	})

	log.Println("🔐 Auth Service running on port 3002")
	app.Listen(":3002")
}