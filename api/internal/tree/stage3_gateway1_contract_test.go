package tree

import (
	"reflect"
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
