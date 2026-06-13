package assets

import (
	"reflect"
	"testing"
)

func TestStage3Gateway1AssetsExposeDraftPublishedIsolation(t *testing.T) {
	assetType := reflect.TypeOf(FileAsset{})
	if _, ok := assetType.FieldByName("State"); !ok {
		t.Fatalf("FileAsset must expose State so Author UI can distinguish draft, published, and draft_and_published assets")
	}

	repoType := reflect.TypeOf((*Repository)(nil)).Elem()
	for _, method := range []string{"ListAssetState", "FindDraftAsset", "PromoteDraftAssets"} {
		if _, ok := repoType.MethodByName(method); !ok {
			t.Fatalf("assets.Repository missing Stage 3 method %s for draft/published asset isolation", method)
		}
	}
}
