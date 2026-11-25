package tests

import (
	"Squirm/back_end/database"
	"Squirm/back_end/handlers"
	"Squirm/back_end/repository"
	"Squirm/back_end/router"
	"Squirm/back_end/services"
	"errors"
	"fmt"

	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Initialize the test database.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connectionString := os.Getenv("TEST_DATABASE_URL")

	if connectionString == "" {
		t.Fatal("TEST_DATABASE_URL is not configured")
	}

	db, err := database.InitDB(
		context.Background(),
		connectionString,
	)
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Prevent accidentally testing against the development DB.
	var databaseName string

	err = db.QueryRowContext(
		context.Background(),
		"SELECT current_database()",
	).Scan(&databaseName)

	if err != nil {
		db.Close()
		t.Fatalf("failed to verify test database: %v", err)
	}

	if databaseName != "squirm_test" {
		db.Close()
		t.Fatalf(
			"expected squirm_test, connected to %s",
			databaseName,
		)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// Initialize the test router and dependencies.
func setupTestRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)

	chatroomRepo := repository.NewChatroomRepository(db)
	chatroomService := services.NewChatroomService(chatroomRepo)
	chatroomHandler := handlers.NewChatroomHandler(chatroomService)

	gin.SetMode(gin.TestMode)

	ginRouter := gin.New()

	// Temporary authentication for testing.
	testAuth := func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Next()
	}

	router.RegisterChatroomRoutes(
		ginRouter,
		chatroomHandler,
		testAuth,
	)

	return ginRouter, db
}

// Helper for executing HTTP requests.
func executeRequest(
	r *gin.Engine,
	method string,
	path string,
	body []byte,
) *httptest.ResponseRecorder {

	request := httptest.NewRequest(
		method,
		path,
		bytes.NewReader(body),
	)

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	return response
}

func createTestChatroom(
	t *testing.T,
	r *gin.Engine,
	db *sql.DB,
) int64 {
	t.Helper()

	body := []byte(`{
		"name": "Test Chatroom",
		"description": "Automated testing"
	}`)

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms",
		body,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create test chatroom: status=%d body=%s",
			response.Code,
			response.Body.String(),
		)
	}

	var result struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result.ID <= 0 {
		t.Fatal("invalid generated chatroom ID")
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_, err := db.ExecContext(
			ctx,
			"DELETE FROM chatrooms WHERE id = $1",
			result.ID,
		)

		if err != nil {
			t.Errorf("failed to clean up chatroom %d: %v", result.ID, err)
		}
	})

	return result.ID
}

// Test successful chatroom creation.
func TestGetChatroom(t *testing.T) {
	r, db := setupTestRouter(t)

	chatroomID := createTestChatroom(t, r, db)

	response := executeRequest(
		r,
		http.MethodGet,
		fmt.Sprintf("/chatrooms/%d", chatroomID),
		nil,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	var chatroom struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		OwnerID     int64  `json:"owner_id"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &chatroom); err != nil {
		t.Fatal(err)
	}

	if chatroom.ID != chatroomID {
		t.Errorf("expected ID %d, got %d", chatroomID, chatroom.ID)
	}

	if chatroom.Name != "Test Chatroom" {
		t.Errorf("unexpected name: %s", chatroom.Name)
	}

	if chatroom.OwnerID != 1 {
		t.Errorf("expected owner ID 1, got %d", chatroom.OwnerID)
	}

	if chatroom.Description != "Automated testing" {
		t.Errorf("unexpected description: %s", chatroom.Description)
	}
}

// Test retrieving a nonexistent chatroom.
func TestGetChatroomNotFound(t *testing.T) {
	r, _ := setupTestRouter(t)

	response := executeRequest(
		r,
		http.MethodGet,
		"/chatrooms/999999999",
		nil,
	)

	if response.Code != http.StatusNotFound {
		t.Errorf(
			"expected status 404, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

// Test creating a chatroom with an invalid name.
func TestCreateChatroomInvalidName(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := []byte(`{
        "name": "   ",
        "description": "Invalid chatroom"
    }`)

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms",
		body,
	)

	if response.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status 400, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

func TestListChatrooms(t *testing.T) {
	r, db := setupTestRouter(t)

	firstID := createTestChatroom(t, r, db)
	secondID := createTestChatroom(t, r, db)

	response := executeRequest(
		r,
		http.MethodGet,
		"/chatrooms",
		nil,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	var chatrooms []struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &chatrooms); err != nil {
		t.Fatal(err)
	}

	foundFirst := false
	foundSecond := false

	for _, chatroom := range chatrooms {
		if chatroom.ID == firstID {
			foundFirst = true
		}

		if chatroom.ID == secondID {
			foundSecond = true
		}
	}

	if !foundFirst {
		t.Errorf("chatroom %d missing from listing", firstID)
	}

	if !foundSecond {
		t.Errorf("chatroom %d missing from listing", secondID)
	}
}

func TestDeleteChatroom(t *testing.T) {
	r, db := setupTestRouter(t)

	chatroomID := createTestChatroom(t, r, db)

	response := executeRequest(
		r,
		http.MethodDelete,
		fmt.Sprintf("/chatrooms/%d", chatroomID),
		nil,
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if response.Body.Len() != 0 {
		t.Errorf("expected empty response body for HTTP 204")
	}

	var id int64

	err := db.QueryRowContext(
		context.Background(),
		"SELECT id FROM chatrooms WHERE id = $1",
		chatroomID,
	).Scan(&id)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf(
			"expected chatroom to be deleted, got: %v",
			err,
		)
	}
}

func TestJoinChatroom(t *testing.T) {
	r, db := setupTestRouter(t)

	chatroomID := createTestChatroom(t, r, db)

	response := executeRequest(
		r,
		http.MethodPost,
		fmt.Sprintf("/chatrooms/%d/members", chatroomID),
		nil,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	var count int

	err := db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*)
		 FROM chatroom_members
		 WHERE chatroom_id = $1 AND user_id = $2`,
		chatroomID,
		int64(1),
	).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("expected 1 membership record, got %d", count)
	}
}

func TestJoinChatroomTwice(t *testing.T) {
	r, db := setupTestRouter(t)

	chatroomID := createTestChatroom(t, r, db)

	path := fmt.Sprintf("/chatrooms/%d/members", chatroomID)

	for i := 0; i < 2; i++ {
		response := executeRequest(
			r,
			http.MethodPost,
			path,
			nil,
		)

		if response.Code != http.StatusOK {
			t.Fatalf(
				"join attempt %d: expected 200, got %d: %s",
				i+1,
				response.Code,
				response.Body.String(),
			)
		}
	}

	var count int

	err := db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*)
		 FROM chatroom_members
		 WHERE chatroom_id = $1 AND user_id = $2`,
		chatroomID,
		int64(1),
	).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf(
			"expected exactly 1 membership after joining twice, got %d",
			count,
		)
	}
}

func TestLeaveChatroom(t *testing.T) {
	r, db := setupTestRouter(t)

	chatroomID := createTestChatroom(t, r, db)

	joinResponse := executeRequest(
		r,
		http.MethodPost,
		fmt.Sprintf("/chatrooms/%d/members", chatroomID),
		nil,
	)

	if joinResponse.Code != http.StatusOK {
		t.Fatalf(
			"failed to join chatroom: %s",
			joinResponse.Body.String(),
		)
	}

	leaveResponse := executeRequest(
		r,
		http.MethodDelete,
		fmt.Sprintf("/chatrooms/%d/members/me", chatroomID),
		nil,
	)

	if leaveResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d: %s",
			leaveResponse.Code,
			leaveResponse.Body.String(),
		)
	}

	var count int

	err := db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*)
		 FROM chatroom_members
		 WHERE chatroom_id = $1 AND user_id = $2`,
		chatroomID,
		int64(1),
	).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf(
			"expected membership to be removed, got %d records",
			count,
		)
	}
}

func TestGetChatroomInvalidID(t *testing.T) {
	r, _ := setupTestRouter(t)

	testCases := []struct {
		name string
		id   string
	}{
		{"non-numeric ID", "abc"},
		{"negative ID", "-5"},
		{"zero ID", "0"},
		{"decimal ID", "1.5"},
		{"integer overflow", "999999999999999999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := executeRequest(
				r,
				http.MethodGet,
				"/chatrooms/"+tc.id,
				nil,
			)

			if response.Code != http.StatusBadRequest {
				t.Errorf(
					"expected 400, got %d: %s",
					response.Code,
					response.Body.String(),
				)
			}
		})
	}
}

func TestCreateChatroomMissingName(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := []byte(`{
		"description": "Missing required name"
	}`)

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms",
		body,
	)

	if response.Code != http.StatusBadRequest {
		t.Errorf(
			"expected 400, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

func TestCreateChatroomMalformedJSON(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := []byte(`{
		"name": "Broken JSON",
		"description":
	}`)

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms",
		body,
	)

	if response.Code != http.StatusBadRequest {
		t.Errorf(
			"expected 400, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

func TestJoinNonexistentChatroom(t *testing.T) {
	r, _ := setupTestRouter(t)

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms/999999999/members",
		nil,
	)

	if response.Code != http.StatusNotFound {
		t.Errorf(
			"expected 404, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
}

func setupTestRouterForUser(
	t *testing.T,
	userID int64,
) (*gin.Engine, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)

	chatroomRepo := repository.NewChatroomRepository(db)
	chatroomService := services.NewChatroomService(chatroomRepo)
	chatroomHandler := handlers.NewChatroomHandler(chatroomService)

	gin.SetMode(gin.TestMode)
	ginRouter := gin.New()

	testAuth := func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}

	router.RegisterChatroomRoutes(
		ginRouter,
		chatroomHandler,
		testAuth,
	)

	return ginRouter, db
}

func TestDeleteChatroomUnauthorized(t *testing.T) {
	ownerRouter, db := setupTestRouterForUser(t, 1)

	chatroomID := createTestChatroom(t, ownerRouter, db)

	// Simulate a different authenticated user.
	otherUserRouter, _ := setupTestRouterForUser(t, 2)

	response := executeRequest(
		otherUserRouter,
		http.MethodDelete,
		fmt.Sprintf("/chatrooms/%d", chatroomID),
		nil,
	)

	if response.Code != http.StatusForbidden {
		t.Errorf(
			"expected 403, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	// Verify the chatroom still exists.
	var count int

	err := db.QueryRowContext(
		context.Background(),
		"SELECT COUNT(*) FROM chatrooms WHERE id = $1",
		chatroomID,
	).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf(
			"unauthorized deletion modified database: count=%d",
			count,
		)
	}
}
