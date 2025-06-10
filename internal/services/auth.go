package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cloud-driver/internal/db"
	"cloud-driver/internal/models"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// AuthService provides authentication and user management
type AuthService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewAuthService creates a new authentication service
func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{
		queries: db.New(pool),
		pool:    pool,
	}
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(ctx context.Context, req models.UserRegistrationRequest) (*models.User, error) {
	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// LoginUser authenticates a user and returns a session token
func (s *AuthService) LoginUser(ctx context.Context, req models.UserLoginRequest) (*models.AuthResponse, error) {
	// Get user by username
	user, err := s.queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Generate session token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session with 24 hour expiry
	expiresAt := time.Now().Add(24 * time.Hour)
	expiresAtPG := pgtype.Timestamp{Time: expiresAt, Valid: true}
	session, err := s.queries.CreateUserSession(ctx, db.CreateUserSessionParams{
		UserID:       user.ID,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAtPG,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &models.AuthResponse{
		User: models.User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		},
		SessionToken: session.SessionToken,
		ExpiresAt:    session.ExpiresAt.Time.Format(time.RFC3339),
	}, nil
}

// ValidateSession validates a session token and returns the user
func (s *AuthService) ValidateSession(ctx context.Context, sessionToken string) (*models.User, error) {
	session, err := s.queries.GetUserSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	return &models.User{
		ID:       session.UserID,
		Username: session.Username,
		Email:    session.Email,
	}, nil
}

// LogoutUser invalidates a session
func (s *AuthService) LogoutUser(ctx context.Context, sessionToken string) error {
	return s.queries.DeleteUserSession(ctx, sessionToken)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID int32) (*models.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &models.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	return s.queries.DeleteExpiredSessions(ctx)
}

// generateSessionToken generates a cryptographically secure session token
func (s *AuthService) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
