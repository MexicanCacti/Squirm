package repository

type UserRepository struct {
}

func (userRepository *UserRepository) CreateUser() {
	/*
		Store app-specific user profile
	*/
}

func (userRepository *UserRepository) GetUserByID() {
	/*
		retrieve user w/ Cognito identifier
	*/
}

func (userRepository *UserRepository) GetUserByUsername() {
	/*
		look up user by username
	*/
}

func (userRepository *UserRepository) UpdateUser() {
	/*
		modify permitted profile fields
	*/
}
