package handlers

import (
	"Squirm/back_end/websocket"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	WebSocketManager *websocket.WebSocketManager
}

func NewWebSocketHandler(webSocketManager *websocket.WebSocketManager) *WebSocketHandler {
	return &WebSocketHandler{WebSocketManager: webSocketManager}
}

func (handler *WebSocketHandler) HandleConnection(ginContext *gin.Context) {
	chatroomID, err := strconv.ParseInt(ginContext.Param("id"), 10, 64)
	if err != nil || chatroomID <= 0 {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatroomID"})
		return
	}

	userID, ok := getAuthenticatedUserID(ginContext)
	if !ok {
		return
	}

	err = handler.WebSocketManager.HandleConnection(ginContext.Writer, ginContext.Request, chatroomID, userID)

	if err != nil {
		if ginContext.Writer.Written() {
			log.Printf("WebSocket connection error: %v", err)
			return
		}

		if errors.Is(err, sql.ErrNoRows) {
			ginContext.JSON(http.StatusNotFound, gin.H{"error": "chatroom not found"})
			return
		}

		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "can't establish WebSocket connection"})
		return
	}
}
