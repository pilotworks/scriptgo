package audit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/scriptgo"
)

// StdlibParam represents an actual parameter defined in ScriptGo's TypeScript stdlib.
type StdlibParam struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Optional     bool   `json:"optional,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
}

// StdlibAPIItem represents an actual API symbol defined in ScriptGo's TypeScript stdlib.
type StdlibAPIItem struct {
	Name          string        `json:"name"`
	FullName      string        `json:"full_name"`
	NormalizedKey string        `json:"normalized_key"`
	Kind          string        `json:"kind"` // "function", "method", "class", "constructor", "property"
	Signature     string        `json:"signature"`
	Params        []StdlibParam `json:"params,omitempty"`
	ReturnType    string        `json:"return_type,omitempty"`
	FilePath      string        `json:"file_path,omitempty"`
	LineNumber    int           `json:"line_number,omitempty"`
	Module        string        `json:"module,omitempty"`
}

// StdlibCatalog stores all indexed ScriptGo stdlib API declarations.
type StdlibCatalog struct {
	ItemsByKey    map[string]*StdlibAPIItem
	ItemsByModule map[string][]*StdlibAPIItem
	AllItems      []*StdlibAPIItem
}

// ScanStdlibAPIs scans the embedded ScriptGo TypeScript stdlib files and indexes all declarations.
func ScanStdlibAPIs() (*StdlibCatalog, error) {
	_ = typescriptgo.EnsureStdlib("")

	catalog := &StdlibCatalog{
		ItemsByKey:    make(map[string]*StdlibAPIItem),
		ItemsByModule: make(map[string][]*StdlibAPIItem),
		AllItems:      make([]*StdlibAPIItem, 0),
	}

	embeddedFS := typescriptgo.EmbeddedFS()
	err := fs.WalkDir(embeddedFS, "stdlib", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel("stdlib", path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if strings.HasSuffix(relSlash, ".gitkeep") {
			return nil
		}
		if !strings.HasSuffix(relSlash, ".ts") {
			return nil
		}

		contentBytes, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)
		filePath := filepath.Join("internal", "typescriptgo", "stdlib", relSlash)

		moduleName := strings.TrimSuffix(relSlash, ".ts")
		if strings.HasSuffix(moduleName, ".d") {
			moduleName = strings.TrimSuffix(moduleName, ".d")
		}
		if strings.HasSuffix(moduleName, "/index") {
			moduleName = strings.TrimSuffix(moduleName, "/index")
		}
		if relSlash == "globals.d.ts" {
			moduleName = "globals"
		}

		syntaxFile, err := typescriptgo.ParseFileToSyntax(filePath, content)
		if err != nil {
			return nil
		}

		processSyntaxStatements(catalog, syntaxFile.Statements, moduleName, filePath, content)
		if moduleName == "stream/web" {
			processSyntaxStatements(catalog, syntaxFile.Statements, "webstreams", filePath, content)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk embedded stdlib dir: %w", err)
	}

	return catalog, nil
}

func processSyntaxStatements(catalog *StdlibCatalog, statements []typescriptgo.SyntaxStatement, moduleName, filePath, content string) {
	for _, stmt := range statements {
		switch stmt.Kind {
		case "function", "async_function", "generator_function", "async_generator_function", "declare_function":
			fnName := stmt.Name
			if fnName == "" {
				continue
			}

			lineNum := lineFromOffset(content, stmt.Span.Start)
			params := convertSyntaxParams(stmt.Parameters)
			retType := stmt.Type
			if retType == "" {
				retType = stmt.InferredType
			}

			sig := formatParamSignature(params, retType)
			fullName := fnName
			if moduleName != "globals" {
				fullName = moduleName + "." + fnName
			}

			normKey := strings.ToLower(fullName)
			item := &StdlibAPIItem{
				Name:          fnName,
				FullName:      fullName,
				NormalizedKey: normKey,
				Kind:          "function",
				Signature:     sig,
				Params:        params,
				ReturnType:    retType,
				FilePath:      filePath,
				LineNumber:    lineNum,
				Module:        moduleName,
			}

			registerStdlibItem(catalog, item, moduleName, fnName, "")

		case "class", "declare_class":
			if stmt.Class == nil {
				continue
			}
			cls := stmt.Class
			clsName := cls.Name
			if clsName == "" {
				continue
			}

			clsLineNum := lineFromOffset(content, cls.Span.Start)
			clsFullName := clsName
			if moduleName != "globals" {
				clsFullName = moduleName + "." + clsName
			}

			clsNormKey := strings.ToLower(clsFullName)
			clsItem := &StdlibAPIItem{
				Name:          clsName,
				FullName:      clsFullName,
				NormalizedKey: clsNormKey,
				Kind:          "class",
				Signature:     "class " + clsName,
				FilePath:      filePath,
				LineNumber:    clsLineNum,
				Module:        moduleName,
			}
			registerStdlibItem(catalog, clsItem, moduleName, clsName, "")

			// Constructor
			if cls.Constructor != nil {
				ctorParams := convertSyntaxParams(cls.Constructor.Parameters)
				ctorLine := lineFromOffset(content, cls.Constructor.Span.Start)
				ctorSig := "new " + clsName + formatParamSignature(ctorParams, "")
				ctorFullName := clsFullName + ".constructor"
				ctorItem := &StdlibAPIItem{
					Name:          "constructor",
					FullName:      ctorFullName,
					NormalizedKey: strings.ToLower(ctorFullName),
					Kind:          "constructor",
					Signature:     ctorSig,
					Params:        ctorParams,
					ReturnType:    clsName,
					FilePath:      filePath,
					LineNumber:    ctorLine,
					Module:        moduleName,
				}
				registerStdlibItem(catalog, ctorItem, moduleName, clsName, "constructor")
				clsItem.Params = ctorParams
				clsItem.Signature = ctorSig
			}

			// Methods
			for _, m := range cls.Methods {
				mName := m.Name
				if mName == "" {
					continue
				}
				mLine := lineFromOffset(content, m.Span.Start)
				mParams := convertSyntaxParams(m.Parameters)
				mRetType := m.Type
				if mRetType == "" {
					mRetType = m.InferredType
				}
				mSig := formatParamSignature(mParams, mRetType)

				mFullName := clsFullName + "." + mName
				if !m.IsStatic {
					mFullName = clsFullName + ".prototype." + mName
				}

				mItem := &StdlibAPIItem{
					Name:          mName,
					FullName:      mFullName,
					NormalizedKey: strings.ToLower(mFullName),
					Kind:          "method",
					Signature:     mSig,
					Params:        mParams,
					ReturnType:    mRetType,
					FilePath:      filePath,
					LineNumber:    mLine,
					Module:        moduleName,
				}
				registerStdlibItem(catalog, mItem, moduleName, clsName, mName)
			}

			// Fields / Properties
			for _, f := range cls.Fields {
				fName := f.Name
				if fName == "" {
					continue
				}
				fLine := lineFromOffset(content, f.Span.Start)
				fType := f.Type
				if fType == "" {
					fType = f.InferredType
				}
				fFullName := clsFullName + "." + fName
				if !f.IsStatic {
					fFullName = clsFullName + ".prototype." + fName
				}
				fItem := &StdlibAPIItem{
					Name:          fName,
					FullName:      fFullName,
					NormalizedKey: strings.ToLower(fFullName),
					Kind:          "property",
					Signature:     fName + ": " + fType,
					ReturnType:    fType,
					FilePath:      filePath,
					LineNumber:    fLine,
					Module:        moduleName,
				}
				registerStdlibItem(catalog, fItem, moduleName, clsName, fName)
			}
		}
	}
}

func convertSyntaxParams(parameters []typescriptgo.SyntaxParameter) []StdlibParam {
	res := make([]StdlibParam, 0, len(parameters))
	for _, p := range parameters {
		pType := p.Type
		if pType == "" {
			pType = p.InferredType
		}
		defVal := ""
		if p.Initializer != nil {
			defVal = p.Initializer.Text
		}
		pName := p.Name
		if p.Rest {
			pName = "..." + pName
		}
		res = append(res, StdlibParam{
			Name:         pName,
			Type:         pType,
			Optional:     p.Optional,
			DefaultValue: defVal,
		})
	}
	return res
}

func formatParamSignature(params []StdlibParam, returnType string) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		s := p.Name
		if p.Optional {
			s += "?"
		}
		if p.Type != "" {
			s += ": " + p.Type
		}
		if p.DefaultValue != "" {
			s += " = " + p.DefaultValue
		}
		parts = append(parts, s)
	}
	sig := "(" + strings.Join(parts, ", ") + ")"
	if returnType != "" && returnType != "void" {
		sig += ": " + returnType
	}
	return sig
}

func lineFromOffset(content string, offset int) int {
	if offset <= 0 {
		return 1
	}
	if offset > len(content) {
		offset = len(content)
	}
	lines := 1
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			lines++
		}
	}
	return lines
}

func registerStdlibItem(catalog *StdlibCatalog, item *StdlibAPIItem, moduleName, className, memberName string) {
	catalog.AllItems = append(catalog.AllItems, item)
	catalog.ItemsByModule[moduleName] = append(catalog.ItemsByModule[moduleName], item)

	// Register multiple lookup keys for flexible matching
	catalog.ItemsByKey[item.NormalizedKey] = item

	cleanMod := strings.ToLower(strings.TrimPrefix(moduleName, "node:"))

	if memberName == "" {
		// Top-level function or class
		catalog.ItemsByKey[cleanMod+"."+strings.ToLower(className)] = item
		catalog.ItemsByKey[strings.ToLower(className)] = item
	} else {
		// Method, constructor or field
		lowClass := strings.ToLower(className)
		lowMember := strings.ToLower(memberName)

		catalog.ItemsByKey[cleanMod+"."+lowClass+"."+lowMember] = item
		catalog.ItemsByKey[cleanMod+"."+lowClass+".prototype."+lowMember] = item
		catalog.ItemsByKey[lowClass+"."+lowMember] = item
		catalog.ItemsByKey[lowClass+".prototype."+lowMember] = item
		catalog.ItemsByKey[cleanMod+"."+lowMember] = item
	}
}
