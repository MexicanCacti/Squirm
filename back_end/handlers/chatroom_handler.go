package handlers

import (
	"Squirm/back_end/services"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateChatroomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type ChatroomHandler struct {
	chatroomService *services.ChatroomService
}

func NewChatroomHandler(chatroomService *services.ChatroomService) *ChatroomHandler {
	return &ChatroomHandler{chatroomService: chatroomService}
}

func getAuthenticatedUserID(ginContext *gin.Context) (int64, bool) {
	userID, exists := ginContext.Get("userID")
	if !exists {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return 0, false
	}

	id, ok := userID.(int64)
	if !ok || id <= 0 {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated identity"})
		return 0, false
	}

	return id, true
}

func (chatroomHandler *ChatroomHandler) CreateChatroom(ginContext *gin.Context) {
	var req CreateChatroomRequest
	if err := ginContext.ShouldBindJSON(&req); err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := getAuthenticatedUserID(ginContext)
	if !ok {
		return
	}

	chatroom, err := chatroomHandler.chatroomService.CreateChatroom(ginContext.Request.Context(), req.Name, req.Description, userID)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ginContext.JSON(http.StatusCreated, chatroom)

}

func (chatroomHandler *ChatroomHandler) GetChatroom(ginContext *gin.Context) {
	chatroomID, err := strconv.ParseInt(ginContext.Param("id"), 10, 64)

	if err != nil || chatroomID <= 0 {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatroom ID"})
		return
	}

	chatroom, err := chatroomHandler.chatroomService.GetChatroom(ginContext.Request.Context(), chatroomID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ginContext.JSON(http.StatusNotFound, gin.H{
				"error": "chatroom not found",
			})
			return
		}
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chatroom"})
		return
	}

	ginContext.JSON(http.StatusOK, chatroom)
}

func (chatroomHandler *ChatroomHandler) ListChatrooms(ginContext *gin.Context) {
	chatrooms, err := chatroomHandler.chatroomService.ListChatrooms(ginContext.Request.Context())

	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chatrooms"})
		return
	}
	ginContext.JSON(http.StatusOK, chatrooms)
}

func (chatroomHandler *ChatroomHandler) DeleteChatroom(ginContext *gin.Context) {
	chatroomID, err := strconv.ParseInt(ginContext.Param("id"), 10, 64)
	if err != nil || chatroomID <= 0 {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatroom ID"})
		return
	}

	userID, ok := getAuthenticatedUserID(ginContext)
	if !ok {
		return
	}

	err = chatroomHandler.chatroomService.DeleteChatroom(ginContext.Request.Context(), chatroomID, userID)

	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete chatroom"})
		return
	}
	ginContext.Status(http.StatusNoContent)
}

func (chatroomHandler *ChatroomHandler) JoinChatroom(ginContext *gin.Context) {
	chatroomID, err := strconv.ParseInt(ginContext.Param("id"), 10, 64)
	if err != nil || chatroomID <= 0 {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatroom ID"})
		return
	}

	userID, ok := getAuthenticatedUserID(ginContext)
	if !ok {
		return
	}

	err = chatroomHandler.chatroomService.JoinChatroom(ginContext.Request.Context(), chatroomID, userID)

	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join chatroom"})
		return
	}

	ginContext.JSON(http.StatusOK, gin.H{"message": "successfully joined chatroom"})
}

func (chatroomHandler *ChatroomHandler) LeaveChatroom(ginContext *gin.Context) {
	chatroomID, err := strconv.ParseInt(ginContext.Param("id"), 10, 64)
	if err != nil || chatroomID <= 0 {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatroom ID"})
		return
	}

	userID, ok := getAuthenticatedUserID(ginContext)
	if !ok {
		return
	}

	err = chatroomHandler.chatroomService.LeaveChatroom(ginContext.Request.Context(), chatroomID, userID)

	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave chatroom"})
		return
	}

	ginContext.Status(http.StatusNoContent)
}
