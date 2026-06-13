package comments

import (
	"os"
	"strings"
	"testing"
)

func TestStage3PublishedFileExistsUsesPublishedSnapshotVisibility(t *testing.T) {
	data, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "published_file_contents") || !strings.Contains(source, "pfc.visible") {
		t.Fatalf("PublishedFileExists must use published_file_contents.visible")
	}
	if strings.Contains(source, "fc.status = 'published'") {
		t.Fatalf("PublishedFileExists still gates on mutable file_contents.status")
	}
}
