package services

import (
	"testing"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

func TestIsMatchingVideoFile(t *testing.T) {
	cases := []struct {
		name     string
		file     driver.FileInfo
		expected bool
	}{
		{
			name:     "exact normalized video",
			file:     driver.FileInfo{Name: "MUKD-569.mp4", Type: "mp4"},
			expected: true,
		},
		{
			name:     "prefixed video",
			file:     driver.FileInfo{Name: "xxxx@MUKD-569.mp4", Type: "mp4"},
			expected: true,
		},
		{
			name:     "wrong video",
			file:     driver.FileInfo{Name: "MUKD-570.mp4", Type: "mp4"},
			expected: false,
		},
		{
			name:     "matching non-video",
			file:     driver.FileInfo{Name: "MUKD-569.txt", Type: "txt"},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isMatchingVideoFile(tc.file, "mukd-569")
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestNormalizeVideoMatchName(t *testing.T) {
	cases := map[string]string{
		"MUKD-569ch":              "mukd-569",
		"ROYD-329-C":              "royd-329",
		"FSDSS-894-uncensored-HD": "fsdss-894",
		"358NTR-101":              "ntr-101",
		"358NTR-101ch":            "ntr-101",
		"mukd-569":                "mukd-569",
		"moviech":                 "moviech",
	}

	for input, expected := range cases {
		t.Run(input, func(t *testing.T) {
			got := normalizeVideoMatchName(input)
			if got != expected {
				t.Fatalf("expected %q, got %q", expected, got)
			}
		})
	}
}
