package services

import (
	"Squirm/back_end/repository"
	"Squirm/back_end/types"
	"context"
	"errors"
	"fmt"
	"strings"
)

type ChatroomService struct {
	chatroomRepository *repository.ChatroomRepository
}

func NewChatroomService(chatroomRepository *repository.ChatroomRepository) *ChatroomService {
	return &ChatroomService{chatroomRepository: chatroomRepository}
}

func (chatroomService *ChatroomService) CreateChatroom(ctx context.Context, name string, description string, ownerID int64) (*types.Chatroom, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("chatroom name empty")
	}

	if len(name) > 100 {
		return nil, errors.New("chatroom name too long")
	}

	if ownerID <= 0 {
		return nil, errors.New("invalid ownerID")
	}

	chatroom := &types.Chatroom{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}

	createdChatroom, err := chatroomService.chatroomRepository.CreateChatroom(ctx, chatroom)
	if err != nil {
		return nil, fmt.Errorf("failed to create chatroom %w", err)
	}
	return createdChatroom, nil
}

func (chatroomService *ChatroomService) GetChatroom(ctx context.Context, chatroomID int64) (*types.Chatroom, error) {
	if chatroomID <= 0 {
		return nil, errors.New("invalid chatroomID")
	}

	chatroom, err := chatroomService.chatroomRepository.GetChatroomByID(ctx, chatroomID)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chatroom with id %d: %w", chatroomID, err)
	}

	return chatroom, nil
}

func (chatroomService *ChatroomService) ListChatrooms(ctx context.Context) ([]types.Chatroom, error) {
	chatrooms, err := chatroomService.chatroomRepository.ListChatrooms(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list chatrooms: %w", err)
	}

	return chatrooms, nil
}

func (chatroomService *ChatroomService) DeleteChatroom(ctx context.Context, chatroomID int64, requestingUserID int64) error {
	if chatroomID <= 0 {
		return errors.New("invalid chatroomID")
	}

	chatroom, err := chatroomService.chatroomRepository.GetChatroomByID(ctx, chatroomID)

	if err != nil {
		return fmt.Errorf("failed to get chatroom: %w", err)
	}

	if chatroom.OwnerID != requestingUserID {
		return errors.New("user not authorized to delete chatroom")
	}

	if err := chatroomService.chatroomRepository.DeleteChatroom(ctx, chatroomID); err != nil {
		return fmt.Errorf("failed to delete chatroom: %w", err)
	}

	return nil

}

func (chatroomService *ChatroomService) JoinChatroom(ctx context.Context, chatroomID int64, userID int64) error {
	if chatroomID <= 0 {
		return errors.New("invalid chatroomID")
	}

	if userID <= 0 {
		return errors.New("invalid userID")
	}

	_, err := chatroomService.chatroomRepository.GetChatroomByID(ctx, chatroomID)
	if err != nil {
		return fmt.Errorf("chatroom DNE: %w", err)
	}

	if err := chatroomService.chatroomRepository.AddMember(ctx, chatroomID, userID); err != nil {
		return fmt.Errorf("failed to join chatroom: %w", err)
	}

	return nil
}

func (chatroomService *ChatroomService) LeaveChatroom(ctx context.Context, chatroomID int64, userID int64) error {
	if chatroomID <= 0 {
		return errors.New("invalid chatroomID")
	}

	if userID <= 0 {
		return errors.New("invalid userID")
	}

	if err := chatroomService.chatroomRepository.RemoveMember(ctx, chatroomID, userID); err != nil {
		return fmt.Errorf("failed to leave chatroom: %w", err)
	}

	return nil
}
