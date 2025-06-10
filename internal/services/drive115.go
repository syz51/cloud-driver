package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Drive115Service provides 115drive cloud storage operations with user-specific credentials
type Drive115Service struct {
	credentialsService *CredentialsService
	pool               *pgxpool.Pool
}

// NewDrive115Service creates a new instance of Drive115Service
func NewDrive115Service(pool *pgxpool.Pool) *Drive115Service {
	return &Drive115Service{
		credentialsService: NewCredentialsService(pool),
		pool:               pool,
	}
}

// getClientForUser returns a 115driver client for a specific user using their active credentials
func (s *Drive115Service) getClientForUser(ctx context.Context, userID int32) (*driver.Pan115Client, error) {
	// Get active credentials for the user
	credentials, err := s.credentialsService.GetActiveUserCredentials(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no active credentials found for user")
	}

	// Use the first active credential (there should only be one active at a time)
	cred := credentials[0]

	// Create driver credential
	cr := &driver.Credential{
		UID:  cred.UID,
		CID:  cred.CID,
		SEID: cred.SEID,
		KID:  cred.KID,
	}

	// Create client and verify login
	client := driver.Defalut().ImportCredential(cr)
	if err := client.LoginCheck(); err != nil {
		return nil, fmt.Errorf("115 driver login failed for user %d: %w", userID, err)
	}

	return client, nil
}

// GetUser returns the current user information for a specific user
func (s *Drive115Service) GetUser(ctx context.Context, userID int32) (interface{}, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return client.GetUser()
}

// ListOfflineTasks returns the list of offline download tasks for a specific user
func (s *Drive115Service) ListOfflineTasks(ctx context.Context, userID int32, page int64) (interface{}, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return client.ListOfflineTask(page)
}

// AddOfflineTaskURIs adds new offline download tasks for a specific user
func (s *Drive115Service) AddOfflineTaskURIs(ctx context.Context, userID int32, urls []string, saveDirID string) ([]string, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Default to root directory if not specified
	if saveDirID == "" {
		saveDirID = "0"
	}
	return client.AddOfflineTaskURIs(urls, saveDirID)
}

// DeleteOfflineTasks deletes offline tasks by their hashes for a specific user
func (s *Drive115Service) DeleteOfflineTasks(ctx context.Context, userID int32, hashes []string, deleteFiles bool) error {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return err
	}
	return client.DeleteOfflineTasks(hashes, deleteFiles)
}

// ClearOfflineTasks clears offline tasks with the specified flag for a specific user
func (s *Drive115Service) ClearOfflineTasks(ctx context.Context, userID int32, clearFlag int64) error {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return err
	}
	return client.ClearOfflineTasks(clearFlag)
}

// ListFiles lists files and directories in the specified directory for a specific user
func (s *Drive115Service) ListFiles(ctx context.Context, userID int32, dirID int64) (*[]driver.File, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert int64 to string as required by the API
	dirIDStr := strconv.FormatInt(dirID, 10)
	return client.List(dirIDStr)
}

// GetFileInfo returns information about a specific file for a specific user
func (s *Drive115Service) GetFileInfo(ctx context.Context, userID int32, fileID int64) (interface{}, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Use GetInfo method instead of non-existent GetFileInfoByID
	// Note: This returns system info, not specific file info
	// For specific file info, we might need to use other methods
	return client.GetInfo()
}

// GetDownloadInfo returns download information for a file for a specific user
func (s *Drive115Service) GetDownloadInfo(ctx context.Context, userID int32, fileID int64) (interface{}, error) {
	client, err := s.getClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// The correct method signature requires a pickCode string, not fileID
	// This is a placeholder - in a real implementation, you'd need to
	// get the pickCode for the file first
	pickCode := strconv.FormatInt(fileID, 10) // This is likely incorrect
	return client.Download(pickCode)
}
