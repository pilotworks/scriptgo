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

func TestAuditCompletedModules(t *testing.T) {
	specCacheDir := filepath.Join("..", "..", "testdata", "specs", "nodejs-v22")
	corpusRoot := filepath.Join("..", "compiler", "testdata", "corpus")
	corpus, err := audit.ScanCorpusAPIs(corpusRoot)
	if err != nil {
		t.Fatalf("ScanCorpusAPIs failed: %v", err)
	}

	modules := []string{"console", "stream", "timers", "buffer", "assert", "util", "https", "child_process", "dgram", "http", "readline", "tty", "single-executable-applications", "wasi", "permissions", "repl", "tracing", "module", "modules"}
	for _, modName := range modules {
		doc, err := spec.LoadModuleSpec(specCacheDir, modName)
		if err != nil {
			t.Fatalf("LoadModuleSpec(%s) failed: %v", modName, err)
		}
		report := audit.AuditModule(modName, doc, corpus, nil, nil)
		if report.MissingCount > 0 {
			t.Errorf("Module %s has %d missing APIs (verified: %d/%d)", modName, report.MissingCount, report.VerifiedCount, report.TotalOfficial)
		}
	}
}
