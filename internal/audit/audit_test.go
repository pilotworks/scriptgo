package audit_test

import (
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/audit"
	"github.com/pilotworks/scriptgo/internal/spec"
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

func TestScanStdlibAPIs(t *testing.T) {
	stdlib, err := audit.ScanStdlibAPIs()
	if err != nil {
		t.Fatalf("ScanStdlibAPIs failed: %v", err)
	}

	if len(stdlib.AllItems) == 0 {
		t.Fatalf("expected stdlib items, got 0")
	}

	t.Logf("Found %d stdlib items across %d unique keys and %d modules",
		len(stdlib.AllItems), len(stdlib.ItemsByKey), len(stdlib.ItemsByModule))

	// Verify querystring.unescape
	unescapeItem, ok := stdlib.ItemsByKey["querystring.unescape"]
	if !ok {
		t.Fatalf("expected 'querystring.unescape' in stdlib catalog keys")
	}
	if len(unescapeItem.Params) == 0 || unescapeItem.Params[0].Name != "str" {
		t.Errorf("expected param 'str' in querystring.unescape, got %+v", unescapeItem.Params)
	}

	// Verify string_decoder.stringdecoder or methods
	if _, ok := stdlib.ItemsByKey["string_decoder.stringdecoder.prototype.write"]; !ok {
		if _, ok2 := stdlib.ItemsByKey["stringdecoder.prototype.write"]; !ok2 {
			t.Errorf("expected write method for StringDecoder in catalog")
		}
	}
}

func TestScanTypesNode(t *testing.T) {
	typesCatalog, err := audit.ScanTypesNode("")
	if err != nil {
		t.Logf("ScanTypesNode skipped or returned: %v", err)
		return
	}

	t.Logf("Found %d @types/node items across %d unique keys and %d modules",
		len(typesCatalog.AllItems), len(typesCatalog.ItemsByKey), len(typesCatalog.ItemsByModule))

}

func TestAuditUtil(t *testing.T) {
	specCacheDir := filepath.Join("..", "..", "testdata", "specs", "nodejs-v22")
	corpusRoot := filepath.Join("..", "compiler", "testdata", "corpus")
	doc, err := spec.LoadModuleSpec(specCacheDir, "util")
	if err != nil {
		t.Fatalf("LoadModuleSpec failed: %v", err)
	}
	corpus, err := audit.ScanCorpusAPIs(corpusRoot)
	if err != nil {
		t.Fatalf("ScanCorpusAPIs failed: %v", err)
	}
	report := audit.AuditModule("util", doc, corpus, nil, nil)
	t.Logf("Util Total: %d, Verified: %d, Missing: %d", report.TotalOfficial, report.VerifiedCount, report.MissingCount)
	for _, res := range report.Results {
		t.Logf("Result: key=%s status=%v", res.CanonicalAPI.NormalizedKey, res.Status)
	}
}




