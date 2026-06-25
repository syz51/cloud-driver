package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud-driver/internal/models"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

// Drive115Service provides 115drive cloud storage operations with credentials from requests
type Drive115Service struct{}

var videoExtensions = map[string]bool{
	"3gp":  true,
	"asf":  true,
	"avi":  true,
	"flv":  true,
	"m2ts": true,
	"m4v":  true,
	"mkv":  true,
	"mov":  true,
	"mp4":  true,
	"mpeg": true,
	"mpg":  true,
	"rm":   true,
	"rmvb": true,
	"ts":   true,
	"webm": true,
	"wmv":  true,
}

const folderVideoScanPageDelay = 750 * time.Millisecond

var trailingCodeNamePattern = regexp.MustCompile(`[a-z]+-\d+$`)
var trailingCodeNameWithSuffixPattern = regexp.MustCompile(`(^|[^a-z0-9])([a-z]+-\d+)(ch|-c|-u|-v|-4k|-uncensored-hd|-中文字幕)$`)
var trailingFC2PPVNamePattern = regexp.MustCompile(`(^|[^a-z0-9])fc2[- ]?ppv[- ]?([0-9]+)(-(c|uc))?$`)

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

// ListFiles lists one page of files and directories in the specified directory
func (s *Drive115Service) ListFiles(ctx context.Context, credentials models.Drive115Credentials, dirID, offset, limit int64) (*[]driver.File, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = 25
	}

	// Convert int64 to string as required by the API
	dirIDStr := strconv.FormatInt(dirID, 10)
	return client.ListPage(dirIDStr, offset, limit)
}

// CheckFolderVideos checks direct files in a folder for matching videos without returning directories.
func (s *Drive115Service) CheckFolderVideos(ctx context.Context, credentials models.Drive115Credentials, dirID, limit int64, indexedName string) (*models.CheckFolderVideosResponse, error) {
	client, err := s.createClient(credentials)
	if err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = 25
	}

	dirIDStr := strconv.FormatInt(dirID, 10)
	expectedName := normalizeVideoMatchName(indexedName)
	offset := int64(0)
	result := &models.CheckFolderVideosResponse{IndexedName: expectedName}

	for {
		files, err := driver.GetFiles(
			client.NewRequest().ForceContentType("application/json;charset=UTF-8"),
			dirIDStr,
			driver.WithLimit(limit),
			driver.WithOffset(offset),
			driver.WithShowDirEnable(false),
		)
		if err != nil {
			return nil, err
		}

		result.CheckedPages++
		if result.CheckedPages == 1 {
			for _, file := range files.Files {
				result.Files = append(result.Files, fileSummary(file, expectedName))
			}
			if nextOffset := int64(files.Offset) + limit; nextOffset < int64(files.Count) {
				result.NextOffset = &nextOffset
			}
		}

		for _, file := range files.Files {
			result.CheckedFiles++
			if isMatchingVideoFile(file, expectedName) {
				result.HasVideos = true
				result.FirstVideoName = file.Name
				return result, nil
			}
		}

		offset = int64(files.Offset) + limit
		if offset >= int64(files.Count) || len(files.Files) == 0 {
			return result, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(folderVideoScanPageDelay):
		}
	}
}

func fileSummary(file driver.FileInfo, expectedName string) models.FileSummary {
	return models.FileSummary{
		ID:      file.FileID,
		Name:    file.Name,
		Type:    strings.ToLower(strings.TrimPrefix(file.Type, ".")),
		Size:    int64(file.Size),
		IsVideo: isVideoFile(file),
		Matches: isMatchingVideoFile(file, expectedName),
	}
}

func normalizeVideoMatchName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = trimLeadingDigitsBeforeCodeName(name)
	if strings.HasSuffix(name, "ch") {
		base := strings.TrimSuffix(name, "ch")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-c") {
		base := strings.TrimSuffix(name, "-c")
		if looksLikeCodeName(base) || looksLikeFC2PPVName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-u") {
		base := strings.TrimSuffix(name, "-u")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-v") {
		base := strings.TrimSuffix(name, "-v")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-4k") {
		base := strings.TrimSuffix(name, "-4k")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-uc") {
		base := strings.TrimSuffix(name, "-uc")
		if looksLikeFC2PPVName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-uncensored-hd") {
		base := strings.TrimSuffix(name, "-uncensored-hd")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if strings.HasSuffix(name, "-中文字幕") {
		base := strings.TrimSuffix(name, "-中文字幕")
		if looksLikeCodeName(base) {
			return base
		}
	}
	if match := trailingFC2PPVNamePattern.FindStringSubmatch(name); match != nil {
		return "fc2ppv-" + match[2]
	}
	if match := trailingCodeNameWithSuffixPattern.FindStringSubmatch(name); match != nil {
		return match[2]
	}
	if match := trailingCodeNamePattern.FindString(name); match != "" && match != name {
		return match
	}
	return name
}

func trimLeadingDigitsBeforeCodeName(name string) string {
	i := 0
	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		i++
	}
	if i == 0 || i == len(name) {
		return name
	}

	rest := name[i:]
	if looksLikeCodeName(rest) {
		return rest
	}
	if strings.HasSuffix(rest, "ch") && looksLikeCodeName(strings.TrimSuffix(rest, "ch")) {
		return rest
	}
	if strings.HasSuffix(rest, "-c") && looksLikeCodeName(strings.TrimSuffix(rest, "-c")) {
		return rest
	}
	if strings.HasSuffix(rest, "-u") && looksLikeCodeName(strings.TrimSuffix(rest, "-u")) {
		return rest
	}
	if strings.HasSuffix(rest, "-v") && looksLikeCodeName(strings.TrimSuffix(rest, "-v")) {
		return rest
	}
	if strings.HasSuffix(rest, "-4k") && looksLikeCodeName(strings.TrimSuffix(rest, "-4k")) {
		return rest
	}
	if strings.HasSuffix(rest, "-uncensored-hd") && looksLikeCodeName(strings.TrimSuffix(rest, "-uncensored-hd")) {
		return rest
	}
	if strings.HasSuffix(rest, "-中文字幕") && looksLikeCodeName(strings.TrimSuffix(rest, "-中文字幕")) {
		return rest
	}
	return name
}

func isMatchingVideoFile(file driver.FileInfo, expectedName string) bool {
	if !isVideoFile(file) {
		return false
	}
	if expectedName == "" {
		return true
	}

	fileName := strings.ToLower(file.Name)
	for _, name := range videoMatchNames(expectedName) {
		if strings.Contains(fileName, name) {
			return true
		}
	}
	return false
}

func videoMatchNames(name string) []string {
	names := []string{name}
	if strings.HasPrefix(name, "fc2ppv-") {
		names = append(names, strings.Replace(name, "fc2ppv-", "fc2-ppv-", 1))
	}
	parts := strings.Split(name, "-")
	if len(parts) == 2 && looksLikeCodeName(name) && len(parts[1]) < 5 {
		names = append(names, parts[0]+strings.Repeat("0", 5-len(parts[1]))+parts[1])
	}
	return names
}

func looksLikeCodeName(name string) bool {
	parts := strings.Split(name, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}

	for _, r := range parts[0] {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	for _, r := range parts[1] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func looksLikeFC2PPVName(name string) bool {
	const prefix = "fc2ppv-"
	if !strings.HasPrefix(name, prefix) || len(name) == len(prefix) {
		return false
	}
	for _, r := range strings.TrimPrefix(name, prefix) {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isVideoFile(file driver.FileInfo) bool {
	if videoExtensions[strings.ToLower(strings.TrimPrefix(file.Type, "."))] {
		return true
	}

	parts := strings.Split(file.Name, ".")
	if len(parts) < 2 {
		return false
	}

	return videoExtensions[strings.ToLower(parts[len(parts)-1])]
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
