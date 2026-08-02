package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"cloud-driver/internal/services"
)

const maxUploadSessionTokenSize = 16 << 10

type uploadSessionCodec struct {
	aead cipher.AEAD
	now  func() time.Time
}

func newUploadSessionCodec(secret string) (*uploadSessionCodec, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("upload session secret must be at least 32 characters")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &uploadSessionCodec{aead: aead, now: time.Now}, nil
}

func (c *uploadSessionCodec) encode(session services.UploadSession) (string, error) {
	plain, err := json.Marshal(session)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, plain, nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *uploadSessionCodec) decode(token string, allowExpired bool) (services.UploadSession, error) {
	var session services.UploadSession
	if token == "" || len(token) > maxUploadSessionTokenSize {
		return session, fmt.Errorf("invalid upload session")
	}
	sealed, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(sealed) < c.aead.NonceSize() {
		return session, fmt.Errorf("invalid upload session")
	}
	nonce, ciphertext := sealed[:c.aead.NonceSize()], sealed[c.aead.NonceSize():]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil || json.Unmarshal(plain, &session) != nil {
		return session, fmt.Errorf("invalid upload session")
	}
	now := c.now().Unix()
	if session.ExpiresAt <= now && (!allowExpired || session.ExpiresAt < now-int64((7*24*time.Hour)/time.Second)) {
		return session, fmt.Errorf("upload session expired")
	}
	if session.PartSize != services.UploadPartSize || session.FileSize <= 0 || session.UploadID == "" || session.Bucket == "" || session.Object == "" {
		return session, fmt.Errorf("invalid upload session")
	}
	return session, nil
}
