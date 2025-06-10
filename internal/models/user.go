package models

import "time"

// User represents a user in the system
type User struct {
	ID           int32     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Drive115Credentials represents 115driver credentials for a user
type Drive115Credentials struct {
	ID        int32     `json:"id" db:"id"`
	UserID    int32     `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	UID       string    `json:"uid" db:"uid"`
	CID       string    `json:"cid" db:"cid"`
	SEID      string    `json:"seid" db:"seid"`
	KID       string    `json:"kid" db:"kid"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserSession represents a user session
type UserSession struct {
	ID           int32     `json:"id" db:"id"`
	UserID       int32     `json:"user_id" db:"user_id"`
	SessionToken string    `json:"session_token" db:"session_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserRegistrationRequest represents a user registration request
type UserRegistrationRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserLoginRequest represents a user login request
type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	User         User   `json:"user"`
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}

// Drive115CredentialsRequest represents a request to add/update 115driver credentials
type Drive115CredentialsRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	UID  string `json:"uid" validate:"required"`
	CID  string `json:"cid" validate:"required"`
	SEID string `json:"seid" validate:"required"`
	KID  string `json:"kid" validate:"required"`
}

// Drive115CredentialsResponse represents the response for credentials operations
type Drive115CredentialsResponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Message  string `json:"message"`
}
