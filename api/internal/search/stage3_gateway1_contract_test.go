package search

import (
	"os"
	"strings"
	"testing"
)

func TestStage3Gateway1SearchReadsPublishedContentSnapshot(t *testing.T) {
	data, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "published_file_contents") {
		t.Fatalf("search repository must query independent published_file_contents snapshots, not mutable file_contents")
	}
	if strings.Contains(source, "fc.status = 'published'") {
		t.Fatalf("search repository still gates mutable file_contents.status instead of Published Content visibility")
	}
}
