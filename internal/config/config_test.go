package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadUploadSecurityConfigFromEnvironment(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("CLOUD_DRIVER_UPLOAD_SESSION_SECRET", "test-upload-session-secret-at-least-32-characters")
	t.Setenv("CLOUD_DRIVER_ALLOWED_ORIGINS", "https://drive.example.com, http://localhost:3012")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://drive.example.com", "http://localhost:3012"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, want) {
		t.Fatalf("allowed origins = %#v, want %#v", cfg.AllowedOrigins, want)
	}
}
