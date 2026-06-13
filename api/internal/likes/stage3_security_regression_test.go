package likes

import (
	"os"
	"strings"
	"testing"
)

func TestStage3FileTargetExistsUsesPublishedSnapshotVisibility(t *testing.T) {
	data, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "published_file_contents") || !strings.Contains(source, "pfc.visible") {
		t.Fatalf("FileTargetExists must use published_file_contents.visible")
	}
	if strings.Contains(source, "fc.status = 'published'") {
		t.Fatalf("FileTargetExists still gates on mutable file_contents.status")
	}
}
