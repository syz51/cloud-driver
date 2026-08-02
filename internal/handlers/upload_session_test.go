package handlers

import (
	"strings"
	"testing"
	"time"

	"cloud-driver/internal/models"
	"cloud-driver/internal/services"
)

func TestUploadSessionCodecRejectsTamperingAndExpiry(t *testing.T) {
	codec, err := newUploadSessionCodec("test-upload-session-secret-at-least-32-characters")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0)
	codec.now = func() time.Time { return now }
	session := services.UploadSession{
		Credentials: models.Drive115Credentials{UID: "uid", CID: "cid", SEID: "seid", KID: "kid"},
		DirID:       "0", FileName: "video.mp4", FileSize: 20, SHA1: strings.Repeat("A", 40),
		PartSize: services.UploadPartSize, Bucket: "bucket", Object: "object", UploadID: "upload", ExpiresAt: now.Add(time.Hour).Unix(),
	}
	token, err := codec.encode(session)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := codec.decode(token, false)
	if err != nil || decoded.FileName != session.FileName || decoded.Credentials.SEID != "seid" {
		t.Fatalf("decoded = %+v, err = %v", decoded, err)
	}
	last := token[len(token)-1]
	tampered := token[:len(token)-1] + string(last^1)
	if _, err := codec.decode(tampered, false); err == nil {
		t.Fatal("tampered token accepted")
	}
	codec.now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, err := codec.decode(token, false); err == nil {
		t.Fatal("expired token accepted")
	}
	if _, err := codec.decode(token, true); err != nil {
		t.Fatalf("expired token not accepted for cleanup: %v", err)
	}
}
