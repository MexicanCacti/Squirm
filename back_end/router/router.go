package router

import (
	"Squirm/back_end/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterChatroomRoutes(router *gin.Engine, chatroomHandler *handlers.ChatroomHandler, webSocketHandler *handlers.WebSocketHandler, authenticationMiddleware gin.HandlerFunc) {
	chatrooms := router.Group("/chatrooms")

	chatrooms.GET("", chatroomHandler.ListChatrooms)
	chatrooms.GET("/:id", chatroomHandler.GetChatroom)

	authenticated := chatrooms.Group("")
	authenticated.Use(authenticationMiddleware)

	authenticated.POST("", chatroomHandler.CreateChatroom)
	authenticated.DELETE("/:id", chatroomHandler.DeleteChatroom)
	authenticated.POST("/:id/members", chatroomHandler.JoinChatroom)
	authenticated.DELETE("/:id/members/me", chatroomHandler.LeaveChatroom)

	authenticated.GET("/:id/ws", webSocketHandler.HandleConnection)
}
