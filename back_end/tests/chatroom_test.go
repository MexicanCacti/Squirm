package tests

import (
	"Squirm/back_end/database"
	"Squirm/back_end/handlers"
	"Squirm/back_end/repository"
	"Squirm/back_end/router"
	"Squirm/back_end/services"

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

// Test successful chatroom creation.
func TestCreateChatroom(t *testing.T) {
	r, db := setupTestRouter(t)

	requestBody := map[string]string{
		"name":        "Test Chatroom",
		"description": "Created by an automated test",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	response := executeRequest(
		r,
		http.MethodPost,
		"/chatrooms",
		body,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d: %s",
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

	if err := json.Unmarshal(
		response.Body.Bytes(),
		&chatroom,
	); err != nil {
		t.Fatal(err)
	}

	// Register cleanup immediately after obtaining the ID.
	if chatroom.ID > 0 {
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			_, err := db.ExecContext(
				ctx,
				"DELETE FROM chatrooms WHERE id = $1",
				chatroom.ID,
			)

			if err != nil {
				t.Errorf(
					"failed to clean up chatroom %d: %v",
					chatroom.ID,
					err,
				)
			}
		})
	}

	if chatroom.ID <= 0 {
		t.Error("expected database-generated chatroom ID")
	}

	if chatroom.Name != requestBody["name"] {
		t.Errorf(
			"expected name %q, got %q",
			requestBody["name"],
			chatroom.Name,
		)
	}

	if chatroom.OwnerID != 1 {
		t.Errorf(
			"expected owner ID 1, got %d",
			chatroom.OwnerID,
		)
	}

	if chatroom.Description != requestBody["description"] {
		t.Errorf(
			"expected description %q, got %q",
			requestBody["description"],
			chatroom.Description,
		)
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
