package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Category represents one of the four standard architectural groups.
type Category string

const (
	CatECMAScript Category = "CategoryECMAScript"
	CatWebCompat  Category = "CategoryWebCompat"
	CatNodeGlobal Category = "CategoryNodeGlobal"
	CatNodeModule Category = "CategoryNodeModule"
)

// ParityStatus represents the implementation status of a symbol.
type ParityStatus string

const (
	StatusDone       ParityStatus = "✅ Done"
	StatusInProgress ParityStatus = "⏳ In Progress"
	StatusPlanned    ParityStatus = "📋 Planned"
	StatusExcluded   ParityStatus = "🚫 Excluded"
)

// APIEntry describes a single method, property, or constructor.
type APIEntry struct {
	Name      string
	Signature string
	Callee    string
	Status    ParityStatus
	TestPath  string
}

// ModuleDocConfig configures the generation for a single markdown checklist.
type ModuleDocConfig struct {
	Category       Category
	ModuleName     string
	FeatureName    string // Maps to internal/compiler/testdata/corpus/<FeatureName>
	DisplayName    string
	ImportPath     string
	SpecRefURL     string
	SpecRefTitle   string
	TypeSourceURL  string
	TypeSourcePath string
	GateOracle     string
	Folder         string
	Entries        []APIEntry
}

// Node.js API JSON structures
type NodeAllDoc struct {
	Modules []NodeJSONModule `json:"modules"`
	Globals []NodeJSONGlobal `json:"globals"`
	Miscs   []NodeJSONMisc   `json:"miscs"`
}

type NodeJSONModule struct {
	Name        string             `json:"name"`
	DisplayName string             `json:"displayName"`
	Desc        string             `json:"desc"`
	Methods     []NodeJSONMethod   `json:"methods"`
	Properties  []NodeJSONProperty `json:"properties"`
	Classes     []NodeJSONClass    `json:"classes"`
	Modules     []NodeJSONModule   `json:"modules"`
}

type NodeJSONClass struct {
	Name       string             `json:"name"`
	TextRaw    string             `json:"textRaw"`
	Desc       string             `json:"desc"`
	Methods    []NodeJSONMethod   `json:"methods"`
	Properties []NodeJSONProperty `json:"properties"`
	Classes    []NodeJSONClass    `json:"classes"`
}

type NodeJSONMethod struct {
	Name       string              `json:"name"`
	TextRaw    string              `json:"textRaw"`
	Desc       string              `json:"desc"`
	Signatures []NodeJSONSignature `json:"signatures"`
}

type NodeJSONSignature struct {
	Params []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"params"`
	Return struct {
		Type string `json:"type"`
	} `json:"return"`
}

type NodeJSONProperty struct {
	Name    string `json:"name"`
	TextRaw string `json:"textRaw"`
	Desc    string `json:"desc"`
}

type NodeJSONGlobal struct {
	Name       string             `json:"name"`
	TextRaw    string             `json:"textRaw"`
	Desc       string             `json:"desc"`
	Methods    []NodeJSONMethod   `json:"methods"`
	Properties []NodeJSONProperty `json:"properties"`
	Classes    []NodeJSONClass    `json:"classes"`
}

type NodeJSONMisc struct {
	Name       string             `json:"name"`
	TextRaw    string             `json:"textRaw"`
	Desc       string             `json:"desc"`
	Methods    []NodeJSONMethod   `json:"methods"`
	Properties []NodeJSONProperty `json:"properties"`
	Classes    []NodeJSONClass    `json:"classes"`
}

func cleanHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := strings.TrimSpace(re.ReplaceAllString(s, ""))
	cleaned = strings.TrimPrefix(cleaned, "Class: ")
	cleaned = strings.TrimPrefix(cleaned, "Event: ")
	if idx := strings.Index(cleaned, " Type: {"); idx != -1 {
		cleaned = cleaned[:idx]
	}
	cleaned = strings.Trim(cleaned, "`")
	return strings.TrimSpace(cleaned)
}

func normalizeAPIName(raw string) string {
	raw = cleanHTML(raw)
	raw = strings.Split(raw, "(")[0]
	raw = strings.TrimPrefix(raw, "new ")
	raw = strings.TrimPrefix(raw, "readonly ")
	parts := strings.Split(raw, ".")
	last := parts[len(parts)-1]
	last = strings.Trim(last, "` :;")

	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchAllCap.ReplaceAllString(last, "${1}_${2}")
	snake = strings.ReplaceAll(snake, "_n_a_n", "_nan")
	snake = strings.ReplaceAll(snake, "_u_t_c", "_utc")
	snake = strings.ReplaceAll(snake, "_u_r_l", "_url")
	return strings.ToLower(strings.ReplaceAll(snake, "-", "_"))
}

var corpusFeatureAPIsCache = make(map[string]map[string]bool)
var corpusFeaturePathCache = make(map[string]string)

func getCorpusFeatureAPIs(featureName string) (string, map[string]bool) {
	if apis, ok := corpusFeatureAPIsCache[featureName]; ok {
		return corpusFeaturePathCache[featureName], apis
	}
	apis := make(map[string]bool)
	resolvedPath := ""

	singleTsPath := filepath.Join("internal", "compiler", "testdata", "corpus", "api", featureName+".ts")
	mainTsPath := filepath.Join("internal", "compiler", "testdata", "corpus", "api", featureName, "main.ts")

	targetPath := ""
	if _, err := os.Stat(singleTsPath); err == nil {
		targetPath = singleTsPath
		resolvedPath = filepath.ToSlash(singleTsPath)
	} else if _, err := os.Stat(mainTsPath); err == nil {
		targetPath = mainTsPath
		resolvedPath = filepath.ToSlash(filepath.Dir(mainTsPath)) + "/"
	} else if langTsPath := filepath.Join("internal", "compiler", "testdata", "corpus", "language", featureName+".ts"); func() bool { _, err := os.Stat(langTsPath); return err == nil }() {
		targetPath = filepath.Join("internal", "compiler", "testdata", "corpus", "language", featureName+".ts")
		resolvedPath = filepath.ToSlash(targetPath)
	} else if featureName == "reflect" {
		decPath := filepath.Join("internal", "compiler", "testdata", "corpus", "language", "decorators.ts")
		if _, err := os.Stat(decPath); err == nil {
			targetPath = decPath
			resolvedPath = filepath.ToSlash(decPath)
		}
	}

	if targetPath != "" {
		if content, err := os.ReadFile(targetPath); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if !strings.HasPrefix(trimmed, "//") {
					continue
				}
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
				if strings.HasPrefix(comment, "@api:") || strings.HasPrefix(comment, "@api ") {
					rawAPI := strings.TrimPrefix(comment, "@api:")
					rawAPI = strings.TrimPrefix(rawAPI, "@api ")
					rawAPI = strings.TrimSpace(rawAPI)
					apis[rawAPI] = true
					apis[strings.ToLower(rawAPI)] = true
					clean := cleanHTML(strings.Split(rawAPI, "(")[0])
					clean = strings.TrimPrefix(clean, "new ")
					clean = strings.TrimPrefix(clean, "readonly ")
					clean = strings.TrimSpace(strings.Split(clean, ":")[0])
					clean = strings.TrimSpace(strings.Split(clean, "<")[0])
					apis[clean] = true
					apis[strings.ToLower(clean)] = true
					parts := strings.Split(clean, ".")
					rawExact := strings.Trim(parts[len(parts)-1], "` :;")
					exactName := strings.TrimPrefix(rawExact, "readonly ")
					exactName = strings.TrimSpace(strings.Split(exactName, ":")[0])
					exactName = strings.TrimSpace(strings.Split(exactName, "<")[0])
					if exactName != "" {
						apis[exactName] = true
						apis[strings.ToLower(exactName)] = true
						apis[normalizeAPIName(exactName)] = true
					}
				}
			}
		}
	}
	corpusFeatureAPIsCache[featureName] = apis
	corpusFeaturePathCache[featureName] = resolvedPath
	return resolvedPath, apis
}

// findCorpusAPITest checks if an API test folder exists under corpus/api/<feature>/ and contains any test cases
func findCorpusAPITest(featureName, apiName string) (string, bool) {
	clean := cleanHTML(strings.Split(apiName, "(")[0])
	clean = strings.ReplaceAll(clean, "`", "")
	clean = strings.TrimPrefix(clean, "new ")
	clean = strings.TrimPrefix(clean, "readonly ")
	clean = strings.TrimSpace(strings.Split(clean, ":")[0])
	clean = strings.TrimSpace(strings.Split(clean, "<")[0])
	parts := strings.Split(clean, ".")
	rawExact := strings.Trim(parts[len(parts)-1], "` :;")
	rawExact = strings.ReplaceAll(rawExact, "`", "")
	exactName := strings.TrimPrefix(rawExact, "readonly ")
	exactName = strings.TrimSpace(strings.Split(exactName, ":")[0])
	exactName = strings.TrimSpace(strings.Split(exactName, "<")[0])
	if strings.Contains(exactName, " ") {
		exactName = strings.Fields(exactName)[0]
	}

	candidates := []string{
		clean,
		strings.ToLower(clean),
		exactName,
		strings.ToLower(exactName),
		rawExact,
		strings.ToLower(rawExact),
		normalizeAPIName(apiName),
		apiName,
		strings.ToLower(apiName),
	}
	if exactName == "constructor" || strings.HasPrefix(apiName, "new ") || exactName == "" {
		candidates = append(candidates, "constructor", "new", "create", "basic")
	}

	features := []string{featureName}
	switch featureName {
	case "headers":
		features = append(features, "fetch", "http")
	case "request", "requestinit":
		features = append(features, "request", "requestinit", "fetch", "http")
	case "response", "responseinit":
		features = append(features, "response", "responseinit", "fetch", "http")
	case "textdecoder", "textdecodeoptions", "textdecoderoptions":
		features = append(features, "textdecoder", "textdecodeoptions", "textdecoderoptions", "encoding")
	case "textencoder", "textencoderencodeintoresult":
		features = append(features, "textencoder", "textencoderencodeintoresult", "encoding")
	case "urlsearchparams":
		features = append(features, "url")
	case "performance":
		features = append(features, "perf_hooks")
	case "iteratorresult", "iteratorobject", "iterator", "asynciterator":
		features = append(features, "iterator", "iteratorresult", "iteratorobject")
	case "erroroptions", "typeerror", "rangeerror", "syntaxerror", "referenceerror", "urierror", "evalerror", "aggregateerror", "suppressederror":
		features = append(features, "error", "suppressederror", "syntaxerror")
	}

	// 1. Check if any feature has a consolidated .ts / main.ts with @api: directives
	for _, feat := range features {
		testPath, inlineAPIs := getCorpusFeatureAPIs(feat)
		if len(inlineAPIs) > 0 {
			for _, cand := range candidates {
				if cand != "" && inlineAPIs[cand] {
					return testPath, true
				}
			}
		}
	}

	// 2. Check if subdirectories exist (legacy / unmigrated features)
	for _, cand := range candidates {
		if cand == "" {
			continue
		}
		candDir := filepath.Join("internal", "compiler", "testdata", "corpus", "api", featureName, cand)
		var hasTest bool
		_ = filepath.WalkDir(candDir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d == nil {
				return nil
			}
			if !d.IsDir() && d.Name() == "main.ts" {
				hasTest = true
			}
			return nil
		})
		if hasTest {
			return filepath.ToSlash(candDir) + "/", true
		}
	}

	return "-", false
}

// getBundledLibsDir returns the directory of bundled TypeScript .d.ts files from typescript-go
func getBundledLibsDir() string {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/microsoft/TypeScript/tsc").Output()
	if err != nil {
		return ""
	}
	modDir := strings.TrimSpace(string(out))
	return filepath.Join(modDir, "internal", "bundled", "libs")
}

// discoverAndParseECMAScript dynamically discovers all ECMAScript built-in interfaces from official .d.ts files
func discoverAndParseECMAScript() []ModuleDocConfig {
	libsDir := getBundledLibsDir()
	if libsDir == "" {
		return nil
	}

	// Dynamically discover all standard ECMAScript lib files from official bundled directory
	dtsFiles, err := filepath.Glob(filepath.Join(libsDir, "lib.es*.d.ts"))
	if err != nil || len(dtsFiles) == 0 {
		return nil
	}

	localGlobals := filepath.Join("internal", "typescriptgo", "stdlib", "globals.d.ts")
	if _, statErr := os.Stat(localGlobals); statErr == nil {
		dtsFiles = append(dtsFiles, localGlobals)
	}

	type rawEntry struct {
		name      string
		signature string
	}

	interfaces := make(map[string][]rawEntry)

	for _, filePath := range dtsFiles {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		currentInterface := ""
		inComment := false
		var memberBuffer strings.Builder
		identRegex := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "/*") {
				inComment = true
			}
			if inComment {
				if strings.Contains(line, "*/") {
					inComment = false
				}
				continue
			}
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "*") {
				continue
			}

			if strings.HasPrefix(line, "interface ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					name := strings.Split(parts[1], "<")[0]
					name = strings.Split(name, " ")[0]
					currentInterface = name
					memberBuffer.Reset()
					continue
				}
			} else if strings.HasPrefix(line, "declare namespace ") || strings.HasPrefix(line, "namespace ") {
				parts := strings.Fields(line)
				idx := 1
				if parts[0] == "declare" {
					idx = 2
				}
				if len(parts) > idx {
					name := strings.Split(parts[idx], "<")[0]
					name = strings.Split(name, " ")[0]
					currentInterface = name
					memberBuffer.Reset()
					continue
				}
			}

			if currentInterface != "" {
				if line == "}" || strings.HasPrefix(line, "}") {
					currentInterface = ""
					memberBuffer.Reset()
					continue
				}

				if memberBuffer.Len() > 0 {
					memberBuffer.WriteString(" ")
				}
				memberBuffer.WriteString(line)

				if !strings.HasSuffix(line, ";") {
					continue
				}

				stmt := strings.TrimSuffix(strings.TrimSpace(memberBuffer.String()), ";")
				memberBuffer.Reset()

				// Collapse multiple whitespaces
				spaceRegex := regexp.MustCompile(`\s+`)
				stmt = spaceRegex.ReplaceAllString(stmt, " ")

				var rawName string
				isConstructor := false
				if strings.HasPrefix(stmt, "new(") || strings.HasPrefix(stmt, "new (") || strings.HasPrefix(stmt, "new<") || strings.HasPrefix(stmt, "new <") || strings.HasPrefix(stmt, "new ") {
					isConstructor = true
					rawName = "constructor"
				} else {
					rawName = stmt
					rawName = strings.TrimPrefix(rawName, "readonly ")
					rawName = strings.TrimPrefix(rawName, "declare function ")
					rawName = strings.TrimPrefix(rawName, "function ")
					rawName = strings.Split(rawName, "<")[0]
					rawName = strings.Split(rawName, "(")[0]
					rawName = strings.Split(rawName, ":")[0]
					rawName = strings.TrimSuffix(strings.TrimSpace(rawName), "?")
				}

				if rawName == "" || (!identRegex.MatchString(rawName) && !isConstructor) || rawName == "prototype" {
					continue
				}

				baseName := strings.TrimSuffix(currentInterface, "Constructor")
				if baseName == "namespace" {
					continue
				}
				interfaces[baseName] = append(interfaces[baseName], rawEntry{
					name:      rawName,
					signature: stmt,
				})
			}
		}
		file.Close()
	}

	var configs []ModuleDocConfig
	for baseName, items := range interfaces {
		modName := strings.ToLower(baseName)
		featName := modName
		seen := make(map[string]bool)
		var entries []APIEntry

		for _, item := range items {
			if seen[item.name] {
				continue
			}
			seen[item.name] = true

			cleanSig := strings.TrimPrefix(item.signature, "declare function ")
			cleanSig = strings.TrimPrefix(cleanSig, "function ")
			displayName := fmt.Sprintf("%s.%s", baseName, cleanSig)
			callee := fmt.Sprintf("__%s.%s", featName, item.name)
			if item.name == "constructor" {
				sigArgs := strings.TrimPrefix(strings.TrimPrefix(item.signature, "new"), " ")
				displayName = fmt.Sprintf("new %s%s", baseName, sigArgs)
				callee = fmt.Sprintf("__%s.new", featName)
			}
			testPath, hasTest := findCorpusAPITest(featName, item.name)
			status := StatusPlanned
			if hasTest {
				status = StatusDone
			}

			entries = append(entries, APIEntry{
				Name:      displayName,
				Signature: cleanSig,
				Callee:    callee,
				Status:    status,
				TestPath:  testPath,
			})
		}

		configs = append(configs, ModuleDocConfig{
			Category:       CatECMAScript,
			ModuleName:     modName,
			FeatureName:    featName,
			DisplayName:    baseName,
			ImportPath:     "N/A (Global Scope)",
			SpecRefURL:     fmt.Sprintf("https://tc39.es/ecma262/#sec-%s-objects", modName),
			SpecRefTitle:   fmt.Sprintf("TC39 ECMA-262 %s Specification", baseName),
			TypeSourceURL:  "https://github.com/microsoft/TypeScript/tree/main/src/lib",
			TypeSourcePath: "microsoft/TypeScript lib.es2024.d.ts",
			GateOracle:     "TC39 Test262 Test Suite & TypeScript baselines",
			Folder:         "ecmascript",
			Entries:        entries,
		})
	}

	return configs
}

// discoverAndParseNodeOfficial dynamically discovers all Node.js and Web APIs from official all.json
func discoverAndParseNodeOfficial() ([]ModuleDocConfig, error) {
	url := "https://nodejs.org/docs/latest-v22.x/api/all.json"
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}

	var allDoc NodeAllDoc
	if err := json.Unmarshal(body, &allDoc); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", url, err)
	}

	var configs []ModuleDocConfig

	extractEntries := func(mod NodeJSONModule, featName string) []APIEntry {
		var entries []APIEntry
		seen := make(map[string]bool)

		addEntry := func(raw string, isProp bool) {
			raw = cleanHTML(raw)
			if raw == "" || seen[raw] {
				return
			}
			if strings.Contains(raw, " ") && !strings.Contains(raw, "(") && !strings.Contains(raw, ".") && !strings.Contains(raw, "[") && !strings.Contains(raw, "`") {
				return
			}
			seen[raw] = true

			cleanCallee := strings.Trim(strings.Split(raw, "(")[0], "`")
			testPath, hasTest := findCorpusAPITest(featName, raw)
			status := StatusPlanned
			if hasTest {
				status = StatusDone
			}

			sig := "(...) => any"
			if isProp {
				sig = "any"
			}

			entries = append(entries, APIEntry{
				Name:      raw,
				Signature: sig,
				Callee:    fmt.Sprintf("__%s.%s", featName, cleanCallee),
				Status:    status,
				TestPath:  testPath,
			})
		}

		var visitClass func(cls NodeJSONClass)
		visitClass = func(cls NodeJSONClass) {
			if cls.TextRaw != "" {
				addEntry(cls.TextRaw, false)
			}
			for _, m := range cls.Methods {
				addEntry(m.TextRaw, false)
			}
			for _, p := range cls.Properties {
				addEntry(p.TextRaw, true)
			}
			for _, sub := range cls.Classes {
				visitClass(sub)
			}
		}

		for _, m := range mod.Methods {
			addEntry(m.TextRaw, false)
		}
		for _, p := range mod.Properties {
			addEntry(p.TextRaw, true)
		}
		for _, cls := range mod.Classes {
			visitClass(cls)
		}
		for _, sub := range mod.Modules {
			for _, sm := range sub.Methods {
				addEntry(sm.TextRaw, false)
			}
			for _, sp := range sub.Properties {
				addEntry(sp.TextRaw, true)
			}
			for _, sc := range sub.Classes {
				visitClass(sc)
			}
		}

		return entries
	}

	for _, mod := range allDoc.Modules {
		rawName := strings.ToLower(mod.Name)
		rawName = strings.TrimPrefix(rawName, "modules:_`node:")
		rawName = strings.TrimSuffix(rawName, "`_api")
		rawName = strings.Trim(rawName, "`")
		rawName = strings.ReplaceAll(rawName, " ", "_")
		rawName = strings.ReplaceAll(rawName, "/", "_")
		rawName = strings.ReplaceAll(rawName, ":", "_")

		if rawName == "" || strings.Contains(rawName, "usage") || strings.Contains(rawName, "example") || strings.Contains(rawName, "embedder") || strings.Contains(rawName, "typescript") {
			continue
		}

		featName := rawName
		category := CatNodeModule
		folder := "nodemodule"
		importPath := fmt.Sprintf("node:%s", rawName)

		if strings.Contains(rawName, "web_crypto") || strings.Contains(rawName, "web_streams") || rawName == "fetch" {
			category = CatWebCompat
			folder = "webcompat"
			importPath = "N/A (Global Scope)"
		} else if rawName == "process" || rawName == "console" || rawName == "globals" {
			category = CatNodeGlobal
			folder = "nodeglobal"
			importPath = "N/A (Global Scope)"
		}

		entries := extractEntries(mod, featName)

		displayName := mod.DisplayName
		if displayName == "" {
			displayName = mod.Name
		}

		configs = append(configs, ModuleDocConfig{
			Category:       category,
			ModuleName:     rawName,
			FeatureName:    featName,
			DisplayName:    displayName,
			ImportPath:     importPath,
			SpecRefURL:     fmt.Sprintf("https://nodejs.org/docs/latest-v22.x/api/%s.html", rawName),
			SpecRefTitle:   fmt.Sprintf("Node.js 22 LTS %s Documentation", displayName),
			TypeSourceURL:  "https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node",
			TypeSourcePath: fmt.Sprintf("@types/node/%s.d.ts", rawName),
			GateOracle:     fmt.Sprintf("Node.js 22 LTS test suite (test/parallel/test-%s-*.js)", rawName),
			Folder:         folder,
			Entries:        entries,
		})
	}

	return configs, nil
}

func renderChecklistMarkdown(cfg ModuleDocConfig) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s Implementation Checklist\n\n", cfg.DisplayName))
	sb.WriteString(fmt.Sprintf("> **Category**: `%s`  \n", cfg.Category))
	sb.WriteString(fmt.Sprintf("> **Import Path**: `%s`  \n", cfg.ImportPath))
	sb.WriteString(fmt.Sprintf("> **Specification Reference**: [%s](%s)  \n", cfg.SpecRefTitle, cfg.SpecRefURL))
	sb.WriteString(fmt.Sprintf("> **Type Definition Source**: [%s](%s)  \n", cfg.TypeSourcePath, cfg.TypeSourceURL))
	sb.WriteString(fmt.Sprintf("> **Gate Oracle**: %s\n\n", cfg.GateOracle))
	sb.WriteString("---\n\n")

	// Section 1: Overview
	sb.WriteString("## 1. Overview & Architectural Pipeline\n\n")
	sb.WriteString("Provide a concise technical summary:\n")
	if cfg.ImportPath == "N/A (Global Scope)" {
		sb.WriteString("- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.\n")
	} else {
		sb.WriteString(fmt.Sprintf("- **Scope & Exposure**: Module-scoped symbols imported explicitly via `%s`.\n", cfg.ImportPath))
	}
	sb.WriteString("- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).\n")
	sb.WriteString("- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.\n\n")
	sb.WriteString("---\n\n")

	// Section 2: Parity Status Matrix
	sb.WriteString("## 2. Parity Status Matrix\n\n")
	sb.WriteString("| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |\n")
	sb.WriteString("| :--- | :--- | :--- | :---: | :--- |\n")

	if len(cfg.Entries) == 0 {
		testPath, _ := findCorpusAPITest(cfg.FeatureName, cfg.ModuleName)
		sb.WriteString(fmt.Sprintf("| `%s` | `any` | `__%s` | 📋 Planned | %s |\n", cfg.ModuleName, cfg.ModuleName, testPath))
	} else {
		for _, e := range cfg.Entries {
			name := strings.ReplaceAll(strings.Trim(e.Name, "`"), "|", "\\|")
			sig := strings.ReplaceAll(e.Signature, "|", "\\|")
			callee := strings.ReplaceAll(strings.Trim(e.Callee, "`"), "|", "\\|")
			testPath := e.TestPath
			if testPath == "" {
				testPath = "-"
			}
			if testPath != "-" {
				testPath = fmt.Sprintf("`%s`", testPath)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %s |\n", name, sig, callee, e.Status, testPath))
		}
	}
	sb.WriteString("\n---\n\n")

	// Section 3: Semantic Details & Edge Cases
	sb.WriteString("## 3. Semantic Details & Edge Cases\n\n")
	sb.WriteString("### 3.1. Standard Behaviors\n")
	sb.WriteString("Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).\n\n")
	sb.WriteString("### 3.2. Native Subset Restrictions\n")
	sb.WriteString("Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).\n\n")
	sb.WriteString("### 3.3. Dual-Surface Mapping (if applicable)\n")
	sb.WriteString(fmt.Sprintf("Corpus test cases for `%s` are organized per API under `internal/compiler/testdata/corpus/%s/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.\n\n", cfg.FeatureName, cfg.FeatureName))
	sb.WriteString("---\n\n")

	// Section 4: Step-by-Step Implementation Recipe
	sb.WriteString("## 4. Step-by-Step Implementation Recipe\n\n")
	sb.WriteString("When implementing or extending any symbol in this file, execute the following technical workflow:\n\n")
	sb.WriteString("- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.\n")
	sb.WriteString("- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.\n")
	sb.WriteString("- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).\n")
	sb.WriteString("- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.\n")
	sb.WriteString("- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.\n")
	sb.WriteString(fmt.Sprintf("- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/%s/<api_name>/<test_case>/` with `main.ts` and `run.expected`.\n", cfg.FeatureName))
	sb.WriteString("- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.\n\n")
	sb.WriteString("---\n\n")

	// Section 5: Known Gaps & Future Roadmap
	sb.WriteString("## 5. Known Gaps & Future Roadmap\n\n")
	sb.WriteString("- [ ] Unimplemented overloads or secondary options arguments.\n")
	sb.WriteString("- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.\n")

	return sb.String()
}

func generateDoc(cfg ModuleDocConfig, outBaseDir string) error {
	// Sort entries deterministically
	sort.SliceStable(cfg.Entries, func(i, j int) bool {
		order := func(s ParityStatus) int {
			switch s {
			case StatusDone:
				return 0
			case StatusInProgress:
				return 1
			case StatusPlanned:
				return 2
			case StatusExcluded:
				return 3
			default:
				return 4
			}
		}
		if order(cfg.Entries[i].Status) != order(cfg.Entries[j].Status) {
			return order(cfg.Entries[i].Status) < order(cfg.Entries[j].Status)
		}
		return cfg.Entries[i].Name < cfg.Entries[j].Name
	})

	content := renderChecklistMarkdown(cfg)

	targetDir := filepath.Join(outBaseDir, cfg.Folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", targetDir, err)
	}

	targetFile := filepath.Join(targetDir, cfg.ModuleName+".md")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", targetFile, err)
	}

	fmt.Printf("✅ Generated %s (%d entries)\n", targetFile, len(cfg.Entries))
	return nil
}

func main() {
	allFlag := flag.Bool("all", true, "Generate all categories")
	categoryFlag := flag.String("category", "", "Specific category to generate (ecmascript, webcompat, nodeglobal, nodemodule)")
	moduleFlag := flag.String("module", "", "Specific module to generate")
	outDirFlag := flag.String("out", "docs/nodejs", "Base output directory")
	flag.Parse()

	fmt.Println("Discovering official ECMAScript standards from TypeScript bundled libs...")
	esConfigs := discoverAndParseECMAScript()
	fmt.Printf("Discovered %d ECMAScript standard specifications.\n", len(esConfigs))

	fmt.Println("Discovering official Node.js APIs from official Node.js all.json specification...")
	nodeConfigs, err := discoverAndParseNodeOfficial()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching official Node.js spec: %v\n", err)
	} else {
		fmt.Printf("Discovered %d Node.js modules and globals.\n", len(nodeConfigs))
	}

	var allConfigs []ModuleDocConfig
	allConfigs = append(allConfigs, esConfigs...)
	allConfigs = append(allConfigs, nodeConfigs...)

	filtered := allConfigs
	if *categoryFlag != "" {
		var list []ModuleDocConfig
		for _, c := range filtered {
			if strings.EqualFold(c.Folder, *categoryFlag) {
				list = append(list, c)
			}
		}
		filtered = list
	}

	if *moduleFlag != "" {
		var list []ModuleDocConfig
		for _, c := range filtered {
			if strings.EqualFold(c.ModuleName, *moduleFlag) {
				list = append(list, c)
			}
		}
		filtered = list
	}

	if !*allFlag && *categoryFlag == "" && *moduleFlag == "" {
		fmt.Println("Usage: go run ./scripts/gendocs/main.go [--all] [--category=<cat>] [--module=<mod>] [--out=docs/nodejs]")
		os.Exit(1)
	}

	fmt.Printf("\nGenerating documentation for %d checklist(s) under %s...\n\n", len(filtered), *outDirFlag)

	successCount := 0
	for _, cfg := range filtered {
		if err := generateDoc(cfg, *outDirFlag); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error generating %s/%s: %v\n", cfg.Folder, cfg.ModuleName, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n✨ Successfully generated %d/%d checklists with 0 hardcoding!\n", successCount, len(filtered))
}
