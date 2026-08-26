package audit_test

import (
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/audit"
)

func TestScanCorpusAPIs(t *testing.T) {
	corpusRoot := filepath.Join("..", "compiler", "testdata", "corpus", "api")
	catalog, err := audit.ScanCorpusAPIs(corpusRoot)
	if err != nil {
		t.Fatalf("ScanCorpusAPIs failed: %v", err)
	}

	if len(catalog.AllItems) == 0 {
		t.Fatalf("expected @api annotations in corpus, got 0")
	}

	t.Logf("Found %d corpus @api annotations across %d unique keys and %d modules",
		len(catalog.AllItems), len(catalog.ItemsByKey), len(catalog.ItemsByModule))

	// Verify key presence for timers and assert
	if _, ok := catalog.ItemsByKey["timers.settimeout"]; !ok {
		t.Errorf("expected 'timers.settimeout' in catalog keys")
	}
	if _, ok := catalog.ItemsByKey["assert.ok"]; !ok {
		t.Errorf("expected 'assert.ok' in catalog keys")
	}
}
