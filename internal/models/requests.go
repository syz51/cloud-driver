package models

// OfflineDownloadRequest represents a request to add offline download tasks
type OfflineDownloadRequest struct {
	URLs      []string `json:"urls" validate:"required"`
	SaveDirID string   `json:"save_dir_id"`
}

// TaskListRequest represents a request to list offline tasks
type TaskListRequest struct {
	Page int64 `json:"page"`
}

// DeleteTasksRequest represents a request to delete offline tasks
type DeleteTasksRequest struct {
	Hashes      []string `json:"hashes" validate:"required"`
	DeleteFiles bool     `json:"delete_files"`
}

// ClearTasksRequest represents a request to clear offline tasks
type ClearTasksRequest struct {
	ClearFlag int64 `json:"clear_flag"`
}
