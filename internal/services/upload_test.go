package services

import "testing"

func TestExpectedPartSize(t *testing.T) {
	session := UploadSession{FileSize: UploadPartSize*2 + 7, PartSize: UploadPartSize}
	for part, want := range map[int]int64{1: UploadPartSize, 2: UploadPartSize, 3: 7} {
		got, err := expectedPartSize(session, part)
		if err != nil || got != want {
			t.Fatalf("part %d: got %d, err %v, want %d", part, got, err, want)
		}
	}
	if _, err := expectedPartSize(session, 4); err == nil {
		t.Fatal("out-of-range part accepted")
	}
}
