package services

import "github.com/gin-gonic/gin"

// verify user id & enforce access rules

func DevelopmentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simulate an authenticated user with ID 1.
		c.Set("userID", int64(1))
		c.Next()
	}
}

type AuthenticationService struct {
	// Cognito config
	// User repo
}

func (authenticationService *AuthenticationService) ValidateToken() {
	/*
		Verify JWT token issued by Cognito
		Validate signature, issuer, expiration & intended client
		Extract auth user's Identity
	*/
}

func (authenticationService *AuthenticationService) GetAuthenticatedUser() {
	/*
		Resolve Cognito identity to app user profile
	*/
}

func (authenticationService *AuthenticationService) AuthorizeChatroomAccess() {
	/*
		Check if user can join chatroom
	*/
}

func (authenticationService *AuthenticationService) AuthorizeChatroomModification() {
	/*
		validate whether user can modify/delete chatroom (must be owner)
	*/
}
