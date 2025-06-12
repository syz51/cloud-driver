package models

// Drive115Credentials represents 115driver credentials passed in requests
type Drive115Credentials struct {
	UID  string `json:"uid" validate:"required,drive115_id,min=1,max=100"`
	CID  string `json:"cid" validate:"required,drive115_id,min=1,max=100"`
	SEID string `json:"seid" validate:"required,drive115_id,min=1,max=100"`
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
