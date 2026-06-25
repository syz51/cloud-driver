package services

import (
	"testing"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

func TestIsMatchingVideoFile(t *testing.T) {
	cases := []struct {
		name         string
		expectedName string
		file         driver.FileInfo
		expected     bool
	}{
		{
			name:         "exact normalized video",
			expectedName: "mukd-569",
			file:         driver.FileInfo{Name: "MUKD-569.mp4", Type: "mp4"},
			expected:     true,
		},
		{
			name:         "prefixed video",
			expectedName: "mukd-569",
			file:         driver.FileInfo{Name: "xxxx@MUKD-569.mp4", Type: "mp4"},
			expected:     true,
		},
		{
			name:         "zero padded compact code video",
			expectedName: "savr-1048",
			file:         driver.FileInfo{Name: "4k2.me@savr01048_2_8k.mp4", Type: "mp4"},
			expected:     true,
		},
		{
			name:         "hyphenated fc2 ppv video",
			expectedName: "fc2ppv-4895806",
			file:         driver.FileInfo{Name: "hhd800.com@FC2-PPV-4895806.mp4", Type: "mp4"},
			expected:     true,
		},
		{
			name:         "t38 special video",
			expectedName: "t38-053",
			file:         driver.FileInfo{Name: "T38-053.mp4", Type: "mp4"},
			expected:     true,
		},
		{
			name:         "wrong video",
			expectedName: "mukd-569",
			file:         driver.FileInfo{Name: "MUKD-570.mp4", Type: "mp4"},
			expected:     false,
		},
		{
			name:         "matching non-video",
			expectedName: "mukd-569",
			file:         driver.FileInfo{Name: "MUKD-569.txt", Type: "txt"},
			expected:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isMatchingVideoFile(tc.file, tc.expectedName)
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestNormalizeVideoMatchName(t *testing.T) {
	cases := map[string]string{
		"MUKD-569ch":                            "mukd-569",
		"ROYD-329-C":                            "royd-329",
		"SOE-480-U":                             "soe-480",
		"start-339-v":                           "start-339",
		"SMBD-115-4K":                           "smbd-115",
		"T38-053":                               "t38-053",
		"FSDSS-894-uncensored-HD":               "fsdss-894",
		"MNGS-045-中文字幕":                         "mngs-045",
		"358NTR-101":                            "ntr-101",
		"358NTR-101ch":                          "ntr-101",
		"[7sht.me]IPX-118-C":                    "ipx-118",
		"第一會所新片@SIS001@STCV-595":                "stcv-595",
		"madoubt.com 669659.xyz fc2ppv-4912492": "fc2ppv-4912492",
		"FC2-PPV-4895806":                       "fc2ppv-4895806",
		"FC2PPV-3061625-C":                      "fc2ppv-3061625",
		"FC2PPV-3175924-UC":                     "fc2ppv-3175924",
		"mukd-569":                              "mukd-569",
		"moviech":                               "moviech",
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

func TestVideoMatchNames(t *testing.T) {
	cases := map[string][]string{
		"savr-1048":      {"savr-1048", "savr01048"},
		"mukd-569":       {"mukd-569", "mukd00569"},
		"hodv-22068":     {"hodv-22068"},
		"fc2ppv-4895806": {"fc2ppv-4895806", "fc2-ppv-4895806"},
	}

	for input, expected := range cases {
		t.Run(input, func(t *testing.T) {
			got := videoMatchNames(input)
			if len(got) != len(expected) {
				t.Fatalf("expected %v, got %v", expected, got)
			}
			for i := range expected {
				if got[i] != expected[i] {
					t.Fatalf("expected %v, got %v", expected, got)
				}
			}
		})
	}
}
