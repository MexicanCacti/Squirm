package repository

import (
	"Squirm/back_end/types"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ChatroomRepository struct {
	db *sql.DB
}

func NewChatroomRepository(db *sql.DB) *ChatroomRepository {
	return &ChatroomRepository{db: db}
}

func (chatroomRepository *ChatroomRepository) CreateChatroom(ctx context.Context, chatroom *types.Chatroom) (*types.Chatroom, error) {

	createQuery :=
		`
	INSERT INTO chatrooms(name, owner_id, description)
	VALUES ($1, $2, $3)
	RETURNING id, created_at
	`

	err := chatroomRepository.db.QueryRowContext(ctx, createQuery, chatroom.Name, chatroom.OwnerID, chatroom.Description).Scan(&chatroom.ID, &chatroom.CreatedAt)

	if err != nil {
		return nil, err
	}

	return chatroom, nil
}

func (chatroomRepository *ChatroomRepository) GetChatroomByID(ctx context.Context, id int64) (*types.Chatroom, error) {

	searchQuery :=
		`
	SELECT id, name, owner_id, created_at, description
	FROM chatrooms
	WHERE id = $1
	`

	chatroom := &types.Chatroom{}

	err := chatroomRepository.db.QueryRowContext(ctx, searchQuery, id).Scan(&chatroom.ID, &chatroom.Name, &chatroom.OwnerID, &chatroom.CreatedAt, &chatroom.Description)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Chatroom with ID %d DNE: %w", id, err)
		}
		return nil, err
	}

	return chatroom, nil
}

// Note add pagination later?
func (chatroomRepository *ChatroomRepository) ListChatrooms(ctx context.Context) ([]types.Chatroom, error) {
	searchQuery :=
		`
	SELECT id, name, owner_id, created_at, description FROM chatrooms
	ORDER BY created_at DESC, id DESC
	`

	rows, err := chatroomRepository.db.QueryContext(ctx, searchQuery)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	chatrooms := make([]types.Chatroom, 0)

	for rows.Next() {
		var chatroom types.Chatroom
		err := rows.Scan(&chatroom.ID, &chatroom.Name, &chatroom.OwnerID, &chatroom.CreatedAt, &chatroom.Description)
		if err != nil {
			return nil, err
		}
		chatrooms = append(chatrooms, chatroom)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chatrooms, nil

}

func (chatroomRepository *ChatroomRepository) DeleteChatroom(ctx context.Context, chatroomID int64) error {
	deleteQuery :=
		`
	DELETE FROM chatrooms
	WHERE id = $1
	`

	result, err := chatroomRepository.db.ExecContext(ctx, deleteQuery, chatroomID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (chatroomRepository *ChatroomRepository) AddMember(ctx context.Context, chatroomID int64, userID int64) error {
	insertQuery :=
		`
	INSERT INTO chatroom_members (chatroom_id, user_id)
	VALUES($1, $2)
	ON CONFLICT (chatroom_id, user_id) DO NOTHING
	`
	_, err := chatroomRepository.db.ExecContext(ctx, insertQuery, chatroomID, userID)

	return err
}

func (chatroomRepository *ChatroomRepository) RemoveMember(ctx context.Context, chatroomID int64, userID int64) error {
	deleteQuery :=
		`
	DELETE FROM chatroom_members
	WHERE chatroom_id = $1
	AND user_id = $2
	`

	result, err := chatroomRepository.db.ExecContext(ctx, deleteQuery, chatroomID, userID)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (chatroomRepository *ChatroomRepository) GetMembers(ctx context.Context, chatroomID int64) ([]int64, error) {
	searchQuery :=
		`
	SELECT user_id 
	FROM chatroom_members 
	WHERE chatroom_id = $1 
	ORDER BY joined_at ASC, user_id ASC
	`

	rows, err := chatroomRepository.db.QueryContext(ctx, searchQuery, chatroomID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]int64, 0)

	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}
