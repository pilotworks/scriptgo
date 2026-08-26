package spec_test

import (
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/spec"
)

func TestListModuleAPIs(t *testing.T) {
	cacheDir := filepath.Join("..", "..", "testdata", "specs", "nodejs-v22")
	for _, mod := range []string{"dns", "domain"} {
		doc, err := spec.LoadModuleSpec(cacheDir, mod)
		if err != nil {
			t.Fatalf("LoadModuleSpec failed for %s: %v", mod, err)
		}
		apis := spec.ExtractCanonicalAPIs(mod, doc)
		t.Logf("=== Module: %s (%d APIs) ===", mod, len(apis))
		for i, a := range apis {
			t.Logf("  [%d] Key=%-30s FullName=%-35s Kind=%s", i+1, a.NormalizedKey, a.FullName, a.Kind)
		}
	}
}
