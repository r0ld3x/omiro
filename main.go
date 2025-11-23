package main

import (
	"encoding/json"
	"log"
	"net/http"
	"omiro/middleware"
	"omiro/redis"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	serverID = uuid.NewString()
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	HandshakeTimeout:  10 * time.Second,
	Subprotocols:      []string{"chat"},
	EnableCompression: true,
}

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redis.Init(redis.Config{
		Host:     redisHost,
		Port:     redisPort,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	redis.RegisterServer(serverID)
	redis.StartSignalSubscriber(serverID, deliverToClient)
	e := echo.New()
	e.GET("/ws", func(c echo.Context) error {
		handleWebSocket(c.Response(), c.Request())
		return nil
	})

	e.GET("/session/new", func(c echo.Context) error {
		token, _, err := middleware.GenerateSessionToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session token"})
		}
		return c.JSON(http.StatusOK, map[string]any{"token": token})
	})
	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})
	go redis.StartMatchmaker()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatalf("error starting server: %v", e.Start(":"+port))
}

func deliverToClient(userID string, payload json.RawMessage) {
	clientsMu.RLock()
	client := clients[userID]
	clientsMu.RUnlock()

	if client == nil {
		return
	}

	client.Send <- SendMessageType{
		Type:    websocket.TextMessage,
		Message: payload,
	}
}
