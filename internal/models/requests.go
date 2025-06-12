package models

// Drive115Credentials represents 115driver credentials passed in requests
type Drive115Credentials struct {
	UID  string `json:"uid" validate:"required,drive115_id,min=1,max=100"`
	CID  string `json:"cid" validate:"required,drive115_id,min=1,max=100"`
	SEID string `json:"seid" validate:"required,drive115_id,min=1,max=200"`
	KID  string `json:"kid" validate:"required,drive115_id,min=1,max=100"`
}

// OfflineDownloadRequest represents a request to add offline download tasks
type OfflineDownloadRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	URLs        []string            `json:"urls" validate:"required,min=1,max=50,urls"`
	SaveDirID   string              `json:"save_dir_id" validate:"omitempty,numeric"`
}

// TaskListRequest represents a request to list offline tasks
type TaskListRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	Page        int64               `json:"page" validate:"omitempty,gte=1,lte=1000"`
}

// DeleteTasksRequest represents a request to delete offline tasks
type DeleteTasksRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	Hashes      []string            `json:"hashes" validate:"required,min=1,max=100,dive,min=1"`
	DeleteFiles bool                `json:"delete_files"`
}

// ClearTasksRequest represents a request to clear offline tasks
type ClearTasksRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	ClearFlag   int64               `json:"clear_flag" validate:"omitempty,oneof=0 1 2 3 4 5"`
}

// GetUserRequest represents a request to get user info
type GetUserRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
}

// ListFilesRequest represents a request to list files
type ListFilesRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	DirID       int64               `json:"dir_id" validate:"omitempty,gte=0"`
}

// FileInfoRequest represents a request to get file info
type FileInfoRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	FileID      int64               `json:"file_id" validate:"required,gt=0"`
}

// DownloadRequest represents a request to get download info
type DownloadRequest struct {
	Credentials Drive115Credentials `json:"credentials" validate:"required"`
	FileID      int64               `json:"file_id" validate:"required,gt=0"`
}

// QRCodeStartRequest represents a request to start a QR code session
type QRCodeStartRequest struct {
	// No credentials required for starting a QR session
}

// QRCodeStartResponse represents the response from starting a QR session
type QRCodeStartResponse struct {
	UID           string `json:"uid"`
	Sign          string `json:"sign"`
	Time          int64  `json:"time"`
	QrcodeContent string `json:"qrcode_content"`
}

// QRCodeImageRequest represents a request to get QR code image
type QRCodeImageRequest struct {
	UID string `json:"uid" validate:"required"`
}

// QRCodeStatusRequest represents a request to check QR code scan status
type QRCodeStatusRequest struct {
	UID  string `json:"uid" validate:"required"`
	Sign string `json:"sign" validate:"required"`
	Time int64  `json:"time" validate:"required"`
}

// QRCodeStatusResponse represents the response for QR code status check
type QRCodeStatusResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

// QRCodeLoginRequest represents a request to complete QR code login
type QRCodeLoginRequest struct {
	UID  string `json:"uid" validate:"required"`
	Sign string `json:"sign" validate:"required"`
	Time int64  `json:"time" validate:"required"`
	App  string `json:"app" validate:"omitempty,oneof=web android ios tv alipaymini wechatmini qandroid"`
}

// QRCodeLoginResponse represents the response from completing QR login
type QRCodeLoginResponse struct {
	Credentials Drive115Credentials `json:"credentials"`
	Success     bool                `json:"success"`
	Message     string              `json:"message"`
}
