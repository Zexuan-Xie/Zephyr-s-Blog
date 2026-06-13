package tree

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStage3Gateway1CurrentContentRevisionContract(t *testing.T) {
	inputType := reflect.TypeOf(UpsertFileContentInput{})
	if field, ok := inputType.FieldByName("ExpectedRevision"); !ok {
		t.Fatalf("UpsertFileContentInput must include ExpectedRevision for optimistic concurrency")
	} else if field.Type.Kind() != reflect.Int {
		t.Fatalf("ExpectedRevision type = %s, want int", field.Type)
	}

	contentType := reflect.TypeOf(FileContent{})
	if field, ok := contentType.FieldByName("Revision"); !ok {
		t.Fatalf("FileContent must expose Revision returned by every Current Content save/read")
	} else if field.Type.Kind() != reflect.Int {
		t.Fatalf("Revision type = %s, want int", field.Type)
	}
	if field, ok := contentType.FieldByName("LastSavedAt"); !ok {
		t.Fatalf("FileContent must expose LastSavedAt for autosave status and conflict UI")
	} else if field.Type != reflect.TypeOf(time.Time{}) {
		t.Fatalf("LastSavedAt type = %s, want time.Time", field.Type)
	}
}

func TestStage3Gateway1PublicationModelRepositoryContract(t *testing.T) {
	repoType := reflect.TypeOf((*LifecycleRepository)(nil)).Elem()
	for _, method := range []string{
		"GetFileVersionState",
		"RestorePreviousContent",
		"PublishCurrentSnapshot",
		"PublishedContent",
	} {
		if _, ok := repoType.MethodByName(method); !ok {
			t.Fatalf("LifecycleRepository missing Stage 3 method %s for Current/Previous/Published snapshot model", method)
		}
	}
}

func TestStage3Gateway1MigrationDeclaresVersionAndSnapshotTables(t *testing.T) {
	migration := readStage3MigrationContractSource(t)
	for _, token := range []string{
		"revision",
		"last_saved_at",
		"file_content_previous_versions",
		"published_file_contents",
		"published_asset",
	} {
		if !strings.Contains(migration, token) {
			t.Fatalf("migration contract missing %q for Stage 3 version/snapshot model", token)
		}
	}
}

func TestStage3Gateway1PublicTreeReadsPublishedContentSnapshot(t *testing.T) {
	data, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "published_file_contents") {
		t.Fatalf("public tree repository must read independent published_file_contents snapshots")
	}
	if strings.Contains(source, "fc.status = 'published'") {
		t.Fatalf("public tree repository still gates mutable file_contents.status instead of Published Content visibility")
	}
}

func readStage3MigrationContractSource(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	var combined string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		data, err := os.ReadFile("../../migrations/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		combined += "\n" + string(data)
	}
	return combined
}
