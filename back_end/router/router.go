package router

import (
	"Squirm/back_end/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterChatroomRoutes(router *gin.Engine, handler *handlers.ChatroomHandler, authenticationMiddleware gin.HandlerFunc) {
	chatrooms := router.Group("/chatrooms")
	chatrooms.GET("", handler.ListChatrooms)
	chatrooms.GET("/:id", handler.GetChatroom)
	authenticated := chatrooms.Group("")
	authenticated.Use(authenticationMiddleware)
	authenticated.POST("", handler.CreateChatroom)
	authenticated.DELETE("/:id", handler.DeleteChatroom)
	authenticated.POST("/:id/members", handler.JoinChatroom)
	authenticated.DELETE("/:id/members/me", handler.LeaveChatroom)
}

/*
Responsible for registering the application's HTTP routes.

Methods:
- Register Chatroom Routes
- Register User Routes
- Register Message Routes
- Register WebSocket Routes
- Register Middleware
*/
