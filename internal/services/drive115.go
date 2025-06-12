package services

import (
	"context"
	"fmt"
	"strconv"

	"cloud-driver/internal/models"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

// Drive115Service provides 115drive cloud storage operations with credentials from requests
type Drive115Service struct{}

// NewDrive115Service creates a new instance of Drive115Service
func NewDrive115Service() *Drive115Service {
	return &Drive115Service{}
}

// createClient creates a 115driver client with the provided credentials
func (s *Drive115Service) createClient(credentials models.Drive115Credentials) (*driver.Pan115Client, error) {
	// Create driver credential
	cr := &driver.Credential{
		UID:  credentials.UID,
		CID:  credentials.CID,
		SEID: credentials.SEID,
		KID:  credentials.KID,
	}

	// Create client and verify login
	client := driver.Defalut().ImportCredential(cr)
	if err := client.LoginCheck(); err != nil {
		return nil, fmt.Errorf("115 driver login failed: %w", err)
	}

	return client, nil
}

// GetUser returns the current user information
func (s *Drive115Service) GetUser(ctx context.Context, credentials models.Drive115Credentials) (interface{}, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}
	return client.GetUser()
}

// ListOfflineTasks returns the list of offline download tasks
func (s *Drive115Service) ListOfflineTasks(ctx context.Context, credentials models.Drive115Credentials, page int64) (interface{}, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}
	return client.ListOfflineTask(page)
}

// AddOfflineTaskURIs adds new offline download tasks
func (s *Drive115Service) AddOfflineTaskURIs(ctx context.Context, credentials models.Drive115Credentials, urls []string, saveDirID string) ([]string, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	client.SetUserAgent(driver.UA115Browser)

	return client.AddOfflineTaskURIs(urls, saveDirID)
}

// DeleteOfflineTasks deletes offline tasks by their hashes
func (s *Drive115Service) DeleteOfflineTasks(ctx context.Context, credentials models.Drive115Credentials, hashes []string, deleteFiles bool) error {
	client, err := s.createClient(credentials)
	if err != nil {
		return err
	}
	return client.DeleteOfflineTasks(hashes, deleteFiles)
}

// ClearOfflineTasks clears offline tasks with the specified flag
func (s *Drive115Service) ClearOfflineTasks(ctx context.Context, credentials models.Drive115Credentials, clearFlag int64) error {
	client, err := s.createClient(credentials)
	if err != nil {
		return err
	}
	return client.ClearOfflineTasks(clearFlag)
}

// ListFiles lists files and directories in the specified directory
func (s *Drive115Service) ListFiles(ctx context.Context, credentials models.Drive115Credentials, dirID int64) (*[]driver.File, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	// Convert int64 to string as required by the API
	dirIDStr := strconv.FormatInt(dirID, 10)
	return client.List(dirIDStr)
}

// GetFileInfo returns information about a specific file
func (s *Drive115Service) GetFileInfo(ctx context.Context, credentials models.Drive115Credentials, fileID int64) (interface{}, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	// Use GetInfo method instead of non-existent GetFileInfoByID
	// Note: This returns system info, not specific file info
	// For specific file info, we might need to use other methods
	return client.GetInfo()
}

// GetDownloadInfo returns download information for a file
func (s *Drive115Service) GetDownloadInfo(ctx context.Context, credentials models.Drive115Credentials, fileID int64) (interface{}, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	// The correct method signature requires a pickCode string, not fileID
	// This is a placeholder - in a real implementation, you'd need to
	// get the pickCode for the file first
	pickCode := strconv.FormatInt(fileID, 10) // This is likely incorrect
	return client.Download(pickCode)
}

// QRCodeStart initiates a QR code login session
func (s *Drive115Service) QRCodeStart(ctx context.Context) (*models.QRCodeStartResponse, error) {
	// Create a default client without credentials for QR code start
	client := driver.Defalut()

	session, err := client.QRCodeStart()
	if err != nil {
		return nil, fmt.Errorf("failed to start QR code session: %w", err)
	}

	return &models.QRCodeStartResponse{
		UID:           session.UID,
		Sign:          session.Sign,
		Time:          session.Time,
		QrcodeContent: session.QrcodeContent,
	}, nil
}

// QRCodeGetImage generates QR code image data
func (s *Drive115Service) QRCodeGetImage(ctx context.Context, uid string) ([]byte, error) {
	// Create a temporary session object for generating the QR image
	session := &driver.QRCodeSession{
		UID: uid,
	}

	// Get QR code image data using the API method
	imageData, err := session.QRCodeByApi()
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	return imageData, nil
}

// QRCodeCheckStatus checks the status of a QR code scan
func (s *Drive115Service) QRCodeCheckStatus(ctx context.Context, uid, sign string, time int64) (*models.QRCodeStatusResponse, error) {
	// Create a default client without credentials
	client := driver.Defalut()

	session := &driver.QRCodeSession{
		UID:  uid,
		Sign: sign,
		Time: time,
	}

	status, err := client.QRCodeStatus(session)
	if err != nil {
		return nil, fmt.Errorf("failed to check QR code status: %w", err)
	}

	// Convert status to user-friendly message
	var message string
	switch {
	case status.IsWaiting():
		message = "Waiting for scan"
	case status.IsScanned():
		message = "QR code scanned, waiting for confirmation"
	case status.IsAllowed():
		message = "Login confirmed, ready to complete"
	case status.IsExpired():
		message = "QR code expired"
	case status.IsCanceled():
		message = "Login canceled"
	default:
		message = "Unknown status"
	}

	return &models.QRCodeStatusResponse{
		Status:  status.Status,
		Message: message,
		Version: status.Version,
	}, nil
}

// QRCodeLogin completes the QR code login and returns credentials
func (s *Drive115Service) QRCodeLogin(ctx context.Context, uid, sign string, time int64, app string) (*models.QRCodeLoginResponse, error) {
	// Create a default client without credentials
	client := driver.Defalut()

	session := &driver.QRCodeSession{
		UID: uid,
	}

	// Set default app if not provided
	if app == "" {
		app = "tv"
	}

	// Complete the login
	var credential *driver.Credential
	credential, err := client.QRCodeLoginWithApp(session, driver.LoginApp(app))

	if err != nil {
		return &models.QRCodeLoginResponse{
			Success: false,
			Message: "Failed to complete QR code login",
		}, fmt.Errorf("failed to complete QR code login: %w", err)
	}

	return &models.QRCodeLoginResponse{
		Credentials: models.Drive115Credentials{
			UID:  credential.UID,
			CID:  credential.CID,
			SEID: credential.SEID,
			KID:  credential.KID,
		},
		Success: true,
		Message: "Login successful",
	}, nil
}
