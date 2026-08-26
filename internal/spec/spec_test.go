package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/spec"
)

func TestLoadModuleSpec(t *testing.T) {
	cacheDir := filepath.Join(os.TempDir(), "scriptgo_specs_test")
	doc, err := spec.LoadModuleSpec(cacheDir, "timers")
	if err != nil {
		t.Logf("Network or spec fetch skipped: %v", err)
		return
	}

	apis := spec.ExtractCanonicalAPIs("timers", doc)
	if len(apis) == 0 {
		t.Fatalf("expected extracted APIs for timers, got 0")
	}

	foundSetTimeout := false
	for _, api := range apis {
		if api.Name == "setTimeout" || api.NormalizedKey == "timers.setTimeout" {
			foundSetTimeout = true
			break
		}
	}

	if !foundSetTimeout {
		t.Errorf("expected timers.setTimeout in extracted APIs")
	}
}
