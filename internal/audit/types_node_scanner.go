package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/scriptgo"
)

// ScanTypesNode scans official @types/node declaration files and indexes all TypeScript declarations.
func ScanTypesNode(customDir string) (*StdlibCatalog, error) {
	typesDir := findTypesNodeDir(customDir)
	if typesDir == "" {
		return nil, fmt.Errorf("@types/node directory not found")
	}

	catalog := &StdlibCatalog{
		ItemsByKey:    make(map[string]*StdlibAPIItem),
		ItemsByModule: make(map[string][]*StdlibAPIItem),
		AllItems:      make([]*StdlibAPIItem, 0),
	}

	err := filepath.WalkDir(typesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".d.ts") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(contentBytes)

		relPath, _ := filepath.Rel(typesDir, path)
		moduleName := strings.TrimSuffix(filepath.Base(path), ".d.ts")
		moduleName = strings.TrimPrefix(moduleName, "node:")
		if idx := strings.Index(moduleName, "."); idx != -1 {
			moduleName = moduleName[:idx]
		}

		syntaxFile, err := typescriptgo.ParseFileToSyntax(path, content)
		if err != nil {
			return nil
		}

		displayPath := filepath.Join("@types", "node", relPath)
		processTypesNodeStatements(catalog, syntaxFile.Statements, moduleName, displayPath, content)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return catalog, nil
}

func findTypesNodeDir(customDir string) string {
	if customDir != "" {
		if info, err := os.Stat(customDir); err == nil && info.IsDir() {
			return customDir
		}
	}

	candidates := []string{
		filepath.Join("web", "node_modules", "@types", "node"),
		filepath.Join("..", "web", "node_modules", "@types", "node"),
		filepath.Join("..", "..", "web", "node_modules", "@types", "node"),
		filepath.Join("node_modules", "@types", "node"),
		filepath.Join("..", "node_modules", "@types", "node"),
		filepath.Join("..", "..", "node_modules", "@types", "node"),
	}


	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates, filepath.Join(home, ".cache", "types-node"))
	}

	for _, cand := range candidates {
		if info, err := os.Stat(cand); err == nil && info.IsDir() {
			return cand
		}
	}
	return ""
}

func processTypesNodeStatements(catalog *StdlibCatalog, statements []typescriptgo.SyntaxStatement, moduleName, filePath, content string) {
	for _, stmt := range statements {
		switch stmt.Kind {
		case "namespace":
			nsName := stmt.Name
			if nsName == "" {
				nsName = moduleName
			}
			cleanNS := strings.TrimPrefix(nsName, "node:")
			processTypesNodeStatements(catalog, stmt.Body, cleanNS, filePath, content)

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
			if moduleName != "" && moduleName != "globals" && moduleName != "index" {
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

		case "class", "declare_class", "interface":
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
			if moduleName != "" && moduleName != "globals" && moduleName != "index" && !strings.Contains(clsName, ".") {
				clsFullName = moduleName + "." + clsName
			}

			clsNormKey := strings.ToLower(clsFullName)
			clsItem := &StdlibAPIItem{
				Name:          clsName,
				FullName:      clsFullName,
				NormalizedKey: clsNormKey,
				Kind:          stmt.Kind,
				Signature:     stmt.Kind + " " + clsName,
				FilePath:      filePath,
				LineNumber:    clsLineNum,
				Module:        moduleName,
			}
			registerStdlibItem(catalog, clsItem, moduleName, clsName, "")

			// Constructor on class / interface
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

				if mName == "constructor" && len(mParams) > 0 {
					clsItem.Params = mParams
					clsItem.Signature = "new " + clsName + formatParamSignature(mParams, "")
				}
			}

			// If interface is e.g. ConsoleConstructor -> also register as Console
			if strings.HasSuffix(clsName, "Constructor") && len(clsItem.Params) > 0 {
				baseName := strings.TrimSuffix(clsName, "Constructor")
				if baseName != "" {
					baseFullName := baseName
					if moduleName != "" && moduleName != "globals" && moduleName != "index" && !strings.Contains(baseName, ".") {
						baseFullName = moduleName + "." + baseName
					}
					baseItem := &StdlibAPIItem{
						Name:          baseName,
						FullName:      baseFullName,
						NormalizedKey: strings.ToLower(baseFullName),
						Kind:          "class",
						Signature:     "new " + baseName + formatParamSignature(clsItem.Params, ""),
						Params:        clsItem.Params,
						ReturnType:    baseName,
						FilePath:      filePath,
						LineNumber:    clsLineNum,
						Module:        moduleName,
					}
					registerStdlibItem(catalog, baseItem, moduleName, baseName, "")
				}
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
