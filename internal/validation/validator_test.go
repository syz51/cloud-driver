package validation

import (
	"cloud-driver/internal/models"
	"testing"
)

func TestDrive115CredentialsValidation(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		credentials models.Drive115Credentials
		expectError bool
		errorCount  int
	}{
		{
			name: "Valid credentials",
			credentials: models.Drive115Credentials{
				UID:  "12345",
				CID:  "abcdef",
				SEID: "67890",
				KID:  "xyz123",
			},
			expectError: false,
		},
		{
			name: "Missing UID",
			credentials: models.Drive115Credentials{
				CID:  "abcdef",
				SEID: "67890",
				KID:  "xyz123",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "UID with spaces (invalid)",
			credentials: models.Drive115Credentials{
				UID:  "123 45",
				CID:  "abcdef",
				SEID: "67890",
				KID:  "xyz123",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "All fields empty",
			credentials: models.Drive115Credentials{},
			expectError: true,
			errorCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.credentials)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, but got none")
					return
				}

				if validationErr, ok := err.(ValidationErrors); ok {
					if len(validationErr.Errors) != tt.errorCount {
						t.Errorf("Expected %d validation errors, got %d", tt.errorCount, len(validationErr.Errors))
					}
				} else {
					t.Errorf("Expected ValidationErrors type, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}

func TestOfflineDownloadRequestValidation(t *testing.T) {
	validator := New()

	validCredentials := models.Drive115Credentials{
		UID:  "12345",
		CID:  "abcdef",
		SEID: "67890",
		KID:  "xyz123",
	}

	tests := []struct {
		name        string
		request     models.OfflineDownloadRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: models.OfflineDownloadRequest{
				Credentials: validCredentials,
				URLs:        []string{"https://example.com/file.zip", "http://test.com/file2.zip"},
				SaveDirID:   "12345",
			},
			expectError: false,
		},
		{
			name: "Empty URLs array",
			request: models.OfflineDownloadRequest{
				Credentials: validCredentials,
				URLs:        []string{},
				SaveDirID:   "12345",
			},
			expectError: true,
		},
		{
			name: "Invalid URL format",
			request: models.OfflineDownloadRequest{
				Credentials: validCredentials,
				URLs:        []string{"not-a-url", "https://valid.com"},
				SaveDirID:   "12345",
			},
			expectError: true,
		},
		{
			name: "Too many URLs",
			request: models.OfflineDownloadRequest{
				Credentials: validCredentials,
				URLs:        make([]string, 51), // Exceeds max of 50
				SaveDirID:   "12345",
			},
			expectError: true,
		},
		{
			name: "Valid magnet URL",
			request: models.OfflineDownloadRequest{
				Credentials: validCredentials,
				URLs:        []string{"magnet:?xt=urn:btih:example"},
				SaveDirID:   "12345",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill invalid URLs array with valid URLs for the "too many" test
			if tt.name == "Too many URLs" {
				for i := range tt.request.URLs {
					tt.request.URLs[i] = "https://example.com"
				}
			}

			err := validator.ValidateStruct(tt.request)

			if tt.expectError && err == nil {
				t.Errorf("Expected validation error, but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestTaskListRequestValidation(t *testing.T) {
	validator := New()

	validCredentials := models.Drive115Credentials{
		UID:  "12345",
		CID:  "abcdef",
		SEID: "67890",
		KID:  "xyz123",
	}

	tests := []struct {
		name        string
		request     models.TaskListRequest
		expectError bool
	}{
		{
			name: "Valid request with page",
			request: models.TaskListRequest{
				Credentials: validCredentials,
				Page:        1,
			},
			expectError: false,
		},
		{
			name: "Valid request without page (omitempty)",
			request: models.TaskListRequest{
				Credentials: validCredentials,
				Page:        0,
			},
			expectError: false,
		},
		{
			name: "Page too high",
			request: models.TaskListRequest{
				Credentials: validCredentials,
				Page:        1001,
			},
			expectError: true,
		},
		{
			name: "Negative page",
			request: models.TaskListRequest{
				Credentials: validCredentials,
				Page:        -1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.request)

			if tt.expectError && err == nil {
				t.Errorf("Expected validation error, but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestFileInfoRequestValidation(t *testing.T) {
	validator := New()

	validCredentials := models.Drive115Credentials{
		UID:  "12345",
		CID:  "abcdef",
		SEID: "67890",
		KID:  "xyz123",
	}

	tests := []struct {
		name        string
		request     models.FileInfoRequest
		expectError bool
	}{
		{
			name: "Valid file ID",
			request: models.FileInfoRequest{
				Credentials: validCredentials,
				FileID:      12345,
			},
			expectError: false,
		},
		{
			name: "Zero file ID",
			request: models.FileInfoRequest{
				Credentials: validCredentials,
				FileID:      0,
			},
			expectError: true,
		},
		{
			name: "Negative file ID",
			request: models.FileInfoRequest{
				Credentials: validCredentials,
				FileID:      -1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.request)

			if tt.expectError && err == nil {
				t.Errorf("Expected validation error, but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}
