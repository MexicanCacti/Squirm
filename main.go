package main

import (
	"Squirm/back_end/database"
	"Squirm/back_end/handlers"
	"Squirm/back_end/repository"
	"Squirm/back_end/router"
	"Squirm/back_end/services"
	"Squirm/back_end/websocket"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("connection string not configured")
	}

	db, err := database.InitDB(ctx, connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	chatroomRepository := repository.NewChatroomRepository(db)
	chatroomService := services.NewChatroomService(chatroomRepository)
	chatroomHandler := handlers.NewChatroomHandler(chatroomService)

	chatroomRegistry := websocket.NewChatroomRegistry()
	webSocketManager := websocket.NewWebSocketManager(chatroomRegistry, chatroomService)
	webSocketHandler := handlers.NewWebSocketHandler(webSocketManager)

	ginRouter := gin.Default()

	authMiddleware := services.DevelopmentAuthMiddleware()

	router.RegisterChatroomRoutes(ginRouter, chatroomHandler, webSocketHandler, authMiddleware)

	if err := ginRouter.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
