package services

import (
	"context"
	"fmt"

	"cloud-driver/internal/db"
	"cloud-driver/internal/models"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CredentialsService provides Drive115 credentials management
type CredentialsService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewCredentialsService creates a new credentials service
func NewCredentialsService(pool *pgxpool.Pool) *CredentialsService {
	return &CredentialsService{
		queries: db.New(pool),
		pool:    pool,
	}
}

// AddCredentials adds new Drive115 credentials for a user
func (s *CredentialsService) AddCredentials(ctx context.Context, userID int32, req models.Drive115CredentialsRequest) (*models.Drive115CredentialsResponse, error) {
	// Create the credentials
	credentials, err := s.queries.CreateDrive115Credentials(ctx, db.CreateDrive115CredentialsParams{
		UserID:   userID,
		Name:     req.Name,
		Uid:      req.UID,
		Cid:      req.CID,
		Seid:     req.SEID,
		Kid:      req.KID,
		IsActive: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	return &models.Drive115CredentialsResponse{
		ID:       credentials.ID,
		Name:     credentials.Name,
		IsActive: credentials.IsActive.Bool,
		Message:  "Credentials added successfully",
	}, nil
}

// GetUserCredentials returns all credentials for a user
func (s *CredentialsService) GetUserCredentials(ctx context.Context, userID int32) ([]models.Drive115Credentials, error) {
	credentials, err := s.queries.GetDrive115CredentialsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	result := make([]models.Drive115Credentials, len(credentials))
	for i, cred := range credentials {
		result[i] = models.Drive115Credentials{
			ID:        cred.ID,
			UserID:    cred.UserID,
			Name:      cred.Name,
			UID:       cred.Uid,
			CID:       cred.Cid,
			SEID:      cred.Seid,
			KID:       cred.Kid,
			IsActive:  cred.IsActive.Bool,
			CreatedAt: cred.CreatedAt.Time,
			UpdatedAt: cred.UpdatedAt.Time,
		}
	}

	return result, nil
}

// GetActiveUserCredentials returns only active credentials for a user
func (s *CredentialsService) GetActiveUserCredentials(ctx context.Context, userID int32) ([]models.Drive115Credentials, error) {
	credentials, err := s.queries.GetActiveDrive115CredentialsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user credentials: %w", err)
	}

	result := make([]models.Drive115Credentials, len(credentials))
	for i, cred := range credentials {
		result[i] = models.Drive115Credentials{
			ID:        cred.ID,
			UserID:    cred.UserID,
			Name:      cred.Name,
			UID:       cred.Uid,
			CID:       cred.Cid,
			SEID:      cred.Seid,
			KID:       cred.Kid,
			IsActive:  cred.IsActive.Bool,
			CreatedAt: cred.CreatedAt.Time,
			UpdatedAt: cred.UpdatedAt.Time,
		}
	}

	return result, nil
}

// GetCredentialsByID returns credentials by ID (with ownership check)
func (s *CredentialsService) GetCredentialsByID(ctx context.Context, userID int32, credentialsID int32) (*models.Drive115Credentials, error) {
	credentials, err := s.queries.GetDrive115CredentialsByID(ctx, credentialsID)
	if err != nil {
		return nil, fmt.Errorf("credentials not found: %w", err)
	}

	// Check ownership
	if credentials.UserID != userID {
		return nil, fmt.Errorf("access denied: credentials do not belong to user")
	}

	return &models.Drive115Credentials{
		ID:        credentials.ID,
		UserID:    credentials.UserID,
		Name:      credentials.Name,
		UID:       credentials.Uid,
		CID:       credentials.Cid,
		SEID:      credentials.Seid,
		KID:       credentials.Kid,
		IsActive:  credentials.IsActive.Bool,
		CreatedAt: credentials.CreatedAt.Time,
		UpdatedAt: credentials.UpdatedAt.Time,
	}, nil
}

// UpdateCredentials updates existing credentials
func (s *CredentialsService) UpdateCredentials(ctx context.Context, userID int32, credentialsID int32, req models.Drive115CredentialsRequest) (*models.Drive115CredentialsResponse, error) {
	// First check ownership
	existing, err := s.GetCredentialsByID(ctx, userID, credentialsID)
	if err != nil {
		return nil, err
	}

	// Update the credentials
	updated, err := s.queries.UpdateDrive115Credentials(ctx, db.UpdateDrive115CredentialsParams{
		ID:       credentialsID,
		Name:     req.Name,
		Uid:      req.UID,
		Cid:      req.CID,
		Seid:     req.SEID,
		Kid:      req.KID,
		IsActive: pgtype.Bool{Bool: existing.IsActive, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update credentials: %w", err)
	}

	return &models.Drive115CredentialsResponse{
		ID:       updated.ID,
		Name:     updated.Name,
		IsActive: updated.IsActive.Bool,
		Message:  "Credentials updated successfully",
	}, nil
}

// SetCredentialsActive sets the active status of credentials
func (s *CredentialsService) SetCredentialsActive(ctx context.Context, userID int32, credentialsID int32, isActive bool) error {
	// First check ownership
	_, err := s.GetCredentialsByID(ctx, userID, credentialsID)
	if err != nil {
		return err
	}

	// If setting to active, deactivate all other credentials for this user
	if isActive {
		if err := s.queries.DeactivateAllUserDrive115Credentials(ctx, userID); err != nil {
			return fmt.Errorf("failed to deactivate existing credentials: %w", err)
		}
	}

	// Set the new active status
	return s.queries.SetDrive115CredentialsActive(ctx, db.SetDrive115CredentialsActiveParams{
		ID:       credentialsID,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
}

// DeleteCredentials deletes credentials (with ownership check)
func (s *CredentialsService) DeleteCredentials(ctx context.Context, userID int32, credentialsID int32) error {
	// First check ownership
	_, err := s.GetCredentialsByID(ctx, userID, credentialsID)
	if err != nil {
		return err
	}

	return s.queries.DeleteDrive115Credentials(ctx, credentialsID)
}
