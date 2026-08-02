package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"cloud-driver/internal/models"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const UploadPartSize int64 = 16 << 20

type UploadSession struct {
	Credentials models.Drive115Credentials `json:"credentials"`
	DirID       string                     `json:"dir_id"`
	FileName    string                     `json:"file_name"`
	FileSize    int64                      `json:"file_size"`
	SHA1        string                     `json:"sha1"`
	PartSize    int64                      `json:"part_size"`
	Bucket      string                     `json:"bucket"`
	Object      string                     `json:"object"`
	Callback    string                     `json:"callback"`
	CallbackVar string                     `json:"callback_var"`
	UploadID    string                     `json:"upload_id"`
	ExpiresAt   int64                      `json:"expires_at"`
}

type UploadInitResult struct {
	State     string
	SignKey   string
	SignCheck string
	Session   *UploadSession
}

type UploadProgress struct {
	NextPart      int
	UploadedBytes int64
	Complete      bool
}

func (s *Drive115Service) InitUpload(ctx context.Context, req models.UploadInitRequest, expiresAt int64) (*UploadInitResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := s.createClient(req.Credentials)
	if err != nil {
		return nil, err
	}
	result, err := client.RapidUploadByHash(req.FileSize, req.FileName, req.DirID, req.PreSHA1, req.SHA1, req.SignKey, req.SignValue)
	if err != nil {
		return nil, err
	}

	switch result.Status {
	case 2:
		return &UploadInitResult{State: "instant"}, nil
	case 7:
		return &UploadInitResult{State: "sign_check", SignKey: result.SignKey, SignCheck: result.SignCheck}, nil
	case 1:
		params := result.UploadOSSParams
		uploadID, err := s.beginMultipartUpload(client, &params)
		if err != nil {
			return nil, err
		}
		return &UploadInitResult{
			State: "upload",
			Session: &UploadSession{
				Credentials: req.Credentials,
				DirID:       req.DirID,
				FileName:    req.FileName,
				FileSize:    req.FileSize,
				SHA1:        req.SHA1,
				PartSize:    UploadPartSize,
				Bucket:      params.Bucket,
				Object:      params.Object,
				Callback:    params.Callback.Callback,
				CallbackVar: params.Callback.CallbackVar,
				UploadID:    uploadID,
				ExpiresAt:   expiresAt,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected 115 upload status: %d", result.Status)
	}
}

func (s *Drive115Service) beginMultipartUpload(client *driver.Pan115Client, params *driver.UploadOSSParams) (string, error) {
	bucket, token, err := uploadBucket(client, params.Bucket)
	if err != nil {
		return "", err
	}
	result, err := bucket.InitiateMultipartUpload(
		params.Object,
		oss.SetHeader(driver.OssSecurityTokenHeaderName, token.SecurityToken),
		oss.UserAgentHeader(driver.OSSUserAgent),
		oss.EnableSha1(),
		oss.Sequential(),
	)
	if err != nil {
		return "", err
	}
	return result.UploadID, nil
}

func (s *Drive115Service) UploadPart(ctx context.Context, session UploadSession, partNumber int, source io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	expected, err := expectedPartSize(session, partNumber)
	if err != nil {
		return err
	}
	client := newUploadClient(session.Credentials)
	bucket, token, err := uploadBucket(client, session.Bucket)
	if err != nil {
		return err
	}
	_, err = bucket.UploadPart(
		initResult(session),
		source,
		expected,
		partNumber,
		oss.SetHeader(driver.OssSecurityTokenHeaderName, token.SecurityToken),
		oss.UserAgentHeader(driver.OSSUserAgent),
	)
	return err
}

func (s *Drive115Service) UploadStatus(ctx context.Context, session UploadSession) (*UploadProgress, error) {
	parts, err := s.listUploadParts(ctx, session)
	if err != nil {
		return nil, err
	}
	progress := &UploadProgress{NextPart: 1}
	for index, part := range parts {
		partNumber := index + 1
		if part.PartNumber != partNumber {
			return nil, fmt.Errorf("upload parts are not contiguous at part %d", partNumber)
		}
		expected, err := expectedPartSize(session, partNumber)
		if err != nil || int64(part.Size) != expected {
			return nil, fmt.Errorf("part %d has invalid size", partNumber)
		}
		progress.UploadedBytes += int64(part.Size)
		progress.NextPart++
	}
	progress.Complete = progress.UploadedBytes == session.FileSize
	return progress, nil
}

func (s *Drive115Service) CompleteUpload(ctx context.Context, session UploadSession) error {
	parts, err := s.listUploadParts(ctx, session)
	if err != nil {
		return err
	}
	progress, err := s.UploadStatus(ctx, session)
	if err != nil {
		return err
	}
	if !progress.Complete {
		return fmt.Errorf("upload is incomplete")
	}

	client := newUploadClient(session.Credentials)
	bucket, token, err := uploadBucket(client, session.Bucket)
	if err != nil {
		return err
	}
	completed := make([]oss.UploadPart, len(parts))
	for index, part := range parts {
		completed[index] = oss.UploadPart{PartNumber: part.PartNumber, ETag: part.ETag}
	}
	params := sessionParams(session)
	var callbackBody []byte
	options := append(driver.OssOption(&params, token), oss.CallbackResult(&callbackBody))
	if _, err := bucket.CompleteMultipartUpload(initResult(session), completed, options...); err != nil {
		return err
	}
	var result driver.UploadResult
	if err := json.Unmarshal(callbackBody, &result); err != nil {
		return fmt.Errorf("decode 115 upload callback: %w", err)
	}
	return result.Err(string(callbackBody))
}

func (s *Drive115Service) AbortUpload(ctx context.Context, session UploadSession) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	client := newUploadClient(session.Credentials)
	bucket, token, err := uploadBucket(client, session.Bucket)
	if err != nil {
		return err
	}
	return bucket.AbortMultipartUpload(
		initResult(session),
		oss.SetHeader(driver.OssSecurityTokenHeaderName, token.SecurityToken),
		oss.UserAgentHeader(driver.OSSUserAgent),
	)
}

func (s *Drive115Service) listUploadParts(ctx context.Context, session UploadSession) ([]oss.UploadedPart, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client := newUploadClient(session.Credentials)
	bucket, token, err := uploadBucket(client, session.Bucket)
	if err != nil {
		return nil, err
	}
	options := []oss.Option{
		oss.SetHeader(driver.OssSecurityTokenHeaderName, token.SecurityToken),
		oss.UserAgentHeader(driver.OSSUserAgent),
		oss.MaxParts(1000),
	}
	parts := make([]oss.UploadedPart, 0)
	marker := 0
	for {
		page, err := bucket.ListUploadedParts(initResult(session), append(options, oss.PartNumberMarker(marker))...)
		if err != nil {
			return nil, err
		}
		parts = append(parts, page.UploadedParts...)
		if !page.IsTruncated {
			break
		}
		marker, err = strconv.Atoi(page.NextPartNumberMarker)
		if err != nil || marker <= 0 {
			return nil, fmt.Errorf("invalid next part marker")
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })
	return parts, nil
}

func newUploadClient(credentials models.Drive115Credentials) *driver.Pan115Client {
	return driver.New(driver.UA(driver.UA115Browser)).ImportCredential(&driver.Credential{
		UID: credentials.UID, CID: credentials.CID, SEID: credentials.SEID, KID: credentials.KID,
	})
}

func uploadBucket(client *driver.Pan115Client, bucketName string) (*oss.Bucket, *driver.UploadOSSTokenResp, error) {
	token, err := client.GetOSSToken()
	if err != nil {
		return nil, nil, err
	}
	ossClient, err := oss.New(client.GetOSSEndpoint(false), token.AccessKeyID, token.AccessKeySecret, oss.EnableCRC(true))
	if err != nil {
		return nil, nil, err
	}
	bucket, err := ossClient.Bucket(bucketName)
	return bucket, token, err
}

func initResult(session UploadSession) oss.InitiateMultipartUploadResult {
	return oss.InitiateMultipartUploadResult{Bucket: session.Bucket, Key: session.Object, UploadID: session.UploadID}
}

func sessionParams(session UploadSession) driver.UploadOSSParams {
	params := driver.UploadOSSParams{SHA1: session.SHA1, Bucket: session.Bucket, Object: session.Object}
	params.Callback.Callback = session.Callback
	params.Callback.CallbackVar = session.CallbackVar
	return params
}

func expectedPartSize(session UploadSession, partNumber int) (int64, error) {
	if partNumber < 1 {
		return 0, fmt.Errorf("part number must be positive")
	}
	offset := int64(partNumber-1) * session.PartSize
	if offset >= session.FileSize {
		return 0, fmt.Errorf("part number exceeds file size")
	}
	remaining := session.FileSize - offset
	if remaining < session.PartSize {
		return remaining, nil
	}
	return session.PartSize, nil
}
