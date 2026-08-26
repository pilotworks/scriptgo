package spec

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocRoot is the root object of a Node.js doc JSON file (e.g. timers.json, fs.json).
type DocRoot struct {
	Type    string      `json:"type"`
	Source  string      `json:"source"`
	Modules []DocModule `json:"modules"`
	Globals []DocGlobal `json:"globals,omitempty"`
	Classes []DocClass  `json:"classes,omitempty"`
	Methods []DocMethod `json:"methods,omitempty"`
}

type DocModule struct {
	Name          string        `json:"name"`
	TextRaw       string        `json:"textRaw"`
	Desc          string        `json:"desc"`
	Stability     int           `json:"stability,omitempty"`
	StabilityText string        `json:"stabilityText,omitempty"`
	Classes       []DocClass    `json:"classes,omitempty"`
	Methods       []DocMethod   `json:"methods,omitempty"`
	Properties    []DocProperty `json:"properties,omitempty"`
	Modules       []DocModule   `json:"modules,omitempty"`
}

type DocClass struct {
	Name       string         `json:"name"`
	TextRaw    string         `json:"textRaw"`
	Desc       string         `json:"desc"`
	Signatures []DocSignature `json:"signatures,omitempty"`
	Methods    []DocMethod    `json:"methods,omitempty"`
	Properties []DocProperty  `json:"properties,omitempty"`
	Events     []DocEvent     `json:"events,omitempty"`
}

type DocMethod struct {
	Name          string         `json:"name"`
	TextRaw       string         `json:"textRaw"`
	Desc          string         `json:"desc"`
	Signatures    []DocSignature `json:"signatures,omitempty"`
	Stability     int            `json:"stability,omitempty"`
	StabilityText string         `json:"stabilityText,omitempty"`
}

type DocProperty struct {
	Name    string `json:"name"`
	TextRaw string `json:"textRaw"`
	Desc    string `json:"desc"`
	Type    string `json:"type,omitempty"`
}

type DocEvent struct {
	Name    string `json:"name"`
	TextRaw string `json:"textRaw"`
	Desc    string `json:"desc"`
}

type DocGlobal struct {
	Name    string      `json:"name"`
	TextRaw string      `json:"textRaw"`
	Desc    string      `json:"desc"`
	Methods []DocMethod `json:"methods,omitempty"`
}

type DocSignature struct {
	Params []DocParam `json:"params,omitempty"`
	Return *DocReturn `json:"return,omitempty"`
}

type DocParam struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Desc     string `json:"desc,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

type DocReturn struct {
	Type string `json:"type,omitempty"`
	Desc string `json:"desc,omitempty"`
}

// CanonicalAPI represents a single identifiable API entry in the spec with its full docs and params.
type CanonicalAPI struct {
	Module        string     `json:"module"`
	Class         string     `json:"class,omitempty"`
	Name          string     `json:"name"`
	NormalizedKey string     `json:"normalized_key"` // e.g. "timers.setTimeout", "fs.readFileSync"
	FullName      string     `json:"full_name"`      // e.g. "timers.setTimeout" or "new assert.AssertionError"
	Kind          string     `json:"kind"`           // "method", "property", "class", "event"
	RawSignature  string     `json:"raw_signature"`
	Desc          string     `json:"desc,omitempty"`
	Params        []DocParam `json:"params,omitempty"`
	Return        *DocReturn `json:"return,omitempty"`
	Stability     int        `json:"stability,omitempty"`
	StabilityText string     `json:"stability_text,omitempty"`
}

// CleanSignature strips markdown wrapper backticks from raw signature strings.
func CleanSignature(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, "`", "")
	return s
}

// FormatConstructorSignature builds a canonical constructor signature like "new Console([options])"
// based on structured class metadata and parameter lists.
func FormatConstructorSignature(className string, params []DocParam) string {
	var paramParts []string
	for _, p := range params {
		pName := CleanSignature(p.Name)
		if p.Optional {
			paramParts = append(paramParts, "["+pName+"]")
		} else {
			paramParts = append(paramParts, pName)
		}
	}
	return fmt.Sprintf("new %s(%s)", className, strings.Join(paramParts, ", "))
}

// NormalizeAPIName normalizes various raw names into canonical "module.member" or "Class.member".
func NormalizeAPIName(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.Contains(s, "`") {
		first := strings.Index(s, "`")
		last := strings.Index(s[first+1:], "`")
		if last != -1 {
			inside := strings.TrimSpace(s[first+1 : first+1+last])
			if inside != "" {
				s = inside
			}
		}
	}
	s = strings.ReplaceAll(s, "`", "")
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "type:") || strings.HasPrefix(lower, "type :") || strings.HasPrefix(lower, "return:") || strings.HasPrefix(lower, "returns:") || strings.HasPrefix(lower, "added in:") || strings.HasPrefix(lower, "deprecated in:") {
		return ""
	}
	if idx := strings.Index(s, "("); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimPrefix(s, "Class: ")
	s = strings.TrimPrefix(s, "Event: ")
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")
	if after, ok := strings.CutPrefix(s, "new "); ok {
		s = after
	}
	s = strings.TrimPrefix(s, "node:")
	return strings.TrimSpace(s)
}

func resolveSubModuleName(parentModule, subName string) string {
	cleanSub := NormalizeAPIName(subName)
	cleanParent := strings.TrimPrefix(NormalizeAPIName(parentModule), "node:")
	switch strings.ToLower(cleanSub) {
	case "promises", "strict", "webcrypto", "constants":
		return fmt.Sprintf("%s/%s", cleanParent, cleanSub)
	default:
		return cleanParent
	}
}

// LoadModuleSpec loads official Node.js documentation JSON for a module.
func LoadModuleSpec(cacheDir, moduleName string) (*DocRoot, error) {
	cleanName := strings.TrimPrefix(moduleName, "node:")
	cachedPath := filepath.Join(cacheDir, cleanName+".json")

	// 1. Try reading from cache
	if data, err := os.ReadFile(cachedPath); err == nil {
		var doc DocRoot
		if err := json.Unmarshal(data, &doc); err == nil {
			return &doc, nil
		}
	}

	// 2. Fetch from nodejs.org
	url := fmt.Sprintf("https://nodejs.org/docs/latest-v22.x/api/%s.json", cleanName)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch official spec for %s: %w", moduleName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("official spec HTTP %d for %s", resp.StatusCode, moduleName)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response for %s: %w", moduleName, err)
	}

	var doc DocRoot
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON spec for %s: %w", moduleName, err)
	}

	// Cache to disk if directory exists or can be created
	if err := os.MkdirAll(cacheDir, 0o755); err == nil {
		_ = os.WriteFile(cachedPath, data, 0o644)
	}

	return &doc, nil
}

func getMethodParamsAndReturn(m DocMethod) ([]DocParam, *DocReturn) {
	if len(m.Signatures) > 0 {
		return m.Signatures[0].Params, m.Signatures[0].Return
	}
	return nil, nil
}

func getClassConstructorParams(c DocClass) ([]DocParam, *DocReturn) {
	if len(c.Signatures) > 0 {
		return c.Signatures[0].Params, c.Signatures[0].Return
	}
	return nil, nil
}

func extractPropertyName(p DocProperty) string {
	name := NormalizeAPIName(p.Name)
	if name != "" && !strings.EqualFold(name, "Type") {
		return name
	}
	raw := strings.TrimSpace(p.TextRaw)
	if idx1 := strings.Index(raw, "`"); idx1 != -1 {
		if idx2 := strings.Index(raw[idx1+1:], "`"); idx2 != -1 {
			candidate := strings.TrimSpace(raw[idx1+1 : idx1+1+idx2])
			candidate = NormalizeAPIName(candidate)
			if candidate != "" && !strings.EqualFold(candidate, "Type") {
				return candidate
			}
		}
	}
	if idx := strings.Index(strings.ToLower(raw), "type:"); idx != -1 {
		candidate := strings.TrimSpace(raw[:idx])
		candidate = strings.Trim(candidate, "` :")
		candidate = NormalizeAPIName(candidate)
		if candidate != "" && !strings.EqualFold(candidate, "Type") {
			return candidate
		}
	}
	return NormalizeAPIName(raw)
}

// ExtractCanonicalAPIs flattens a DocRoot into a slice of CanonicalAPI items.
func ExtractCanonicalAPIs(moduleName string, doc *DocRoot) []CanonicalAPI {
	var apis []CanonicalAPI
	cleanModule := strings.TrimPrefix(moduleName, "node:")
	seen := make(map[string]bool)

	addAPI := func(api CanonicalAPI) {
		key := strings.ToLower(api.NormalizedKey)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		apis = append(apis, api)
	}

	var extractFromModule func(mod DocModule, currentModName string)
	extractFromModule = func(mod DocModule, currentModName string) {
		modName := resolveSubModuleName(currentModName, mod.Name)

		for _, m := range mod.Methods {
			name := NormalizeAPIName(m.Name)
			if name == "" {
				name = NormalizeAPIName(m.TextRaw)
			}
			if name == "" {
				continue
			}
			params, ret := getMethodParamsAndReturn(m)
			fullName := fmt.Sprintf("%s.%s", modName, name)
			addAPI(CanonicalAPI{
				Module:        modName,
				Name:          name,
				NormalizedKey: fullName,
				FullName:      fullName,
				Kind:          "method",
				RawSignature:  CleanSignature(m.TextRaw),
				Desc:          m.Desc,
				Params:        params,
				Return:        ret,
				Stability:     m.Stability,
				StabilityText: m.StabilityText,
			})
		}

		for _, p := range mod.Properties {
			name := extractPropertyName(p)
			if name == "" {
				continue
			}
			fullName := fmt.Sprintf("%s.%s", modName, name)
			addAPI(CanonicalAPI{
				Module:        modName,
				Name:          name,
				NormalizedKey: fullName,
				FullName:      fullName,
				Kind:          "property",
				RawSignature:  CleanSignature(p.TextRaw),
				Desc:          p.Desc,
				Return:        &DocReturn{Type: p.Type},
			})
		}

		for _, c := range mod.Classes {
			cName := NormalizeAPIName(c.Name)
			if cName == "" {
				continue
			}
			cParams, cRet := getClassConstructorParams(c)
			addAPI(CanonicalAPI{
				Module:        modName,
				Class:         cName,
				Name:          cName,
				NormalizedKey: fmt.Sprintf("%s.%s", modName, cName),
				FullName:      fmt.Sprintf("new %s.%s", modName, cName),
				Kind:          "class",
				RawSignature:  FormatConstructorSignature(cName, cParams),
				Desc:          c.Desc,
				Params:        cParams,
				Return:        cRet,
			})

			for _, m := range c.Methods {
				name := NormalizeAPIName(m.Name)
				if name == "" {
					continue
				}
				fullName := fmt.Sprintf("%s.%s", cName, name)
				params, ret := getMethodParamsAndReturn(m)
				addAPI(CanonicalAPI{
					Module:        modName,
					Class:         cName,
					Name:          name,
					NormalizedKey: fullName,
					FullName:      fullName,
					Kind:          "method",
					RawSignature:  CleanSignature(m.TextRaw),
					Desc:          m.Desc,
					Params:        params,
					Return:        ret,
					Stability:     m.Stability,
					StabilityText: m.StabilityText,
				})
			}
			for _, p := range c.Properties {
				pName := extractPropertyName(p)
				if pName == "" {
					continue
				}
				fullName := fmt.Sprintf("%s.%s", cName, pName)
				addAPI(CanonicalAPI{
					Module:        modName,
					Class:         cName,
					Name:          pName,
					NormalizedKey: fullName,
					FullName:      fullName,
					Kind:          "property",
					RawSignature:  CleanSignature(p.TextRaw),
					Desc:          p.Desc,
					Return:        &DocReturn{Type: p.Type},
				})
			}
		}

		for _, sub := range mod.Modules {
			extractFromModule(sub, modName)
		}
	}

	for _, mod := range doc.Modules {
		extractFromModule(mod, cleanModule)
	}

	for _, c := range doc.Classes {
		cName := NormalizeAPIName(c.Name)
		if cName == "" {
			continue
		}
		cParams, cRet := getClassConstructorParams(c)
		addAPI(CanonicalAPI{
			Module:        cleanModule,
			Class:         cName,
			Name:          cName,
			NormalizedKey: fmt.Sprintf("%s.%s", cleanModule, cName),
			FullName:      fmt.Sprintf("new %s.%s", cleanModule, cName),
			Kind:          "class",
			RawSignature:  FormatConstructorSignature(cName, cParams),
			Desc:          c.Desc,
			Params:        cParams,
			Return:        cRet,
		})

		for _, m := range c.Methods {
			name := NormalizeAPIName(m.Name)
			if name == "" {
				continue
			}
			fullName := fmt.Sprintf("%s.%s", cName, name)
			params, ret := getMethodParamsAndReturn(m)
			addAPI(CanonicalAPI{
				Module:        cleanModule,
				Class:         cName,
				Name:          name,
				NormalizedKey: fullName,
				FullName:      fullName,
				Kind:          "method",
				RawSignature:  CleanSignature(m.TextRaw),
				Desc:          m.Desc,
				Params:        params,
				Return:        ret,
				Stability:     m.Stability,
				StabilityText: m.StabilityText,
			})
		}
		for _, p := range c.Properties {
			pName := extractPropertyName(p)
			if pName == "" {
				continue
			}
			fullName := fmt.Sprintf("%s.%s", cName, pName)
			addAPI(CanonicalAPI{
				Module:        cleanModule,
				Class:         cName,
				Name:          pName,
				NormalizedKey: fullName,
				FullName:      fullName,
				Kind:          "property",
				RawSignature:  CleanSignature(p.TextRaw),
				Desc:          p.Desc,
				Return:        &DocReturn{Type: p.Type},
			})
		}
	}

	for _, m := range doc.Methods {
		name := NormalizeAPIName(m.Name)
		if name == "" {
			continue
		}
		fullName := fmt.Sprintf("%s.%s", cleanModule, name)
		params, ret := getMethodParamsAndReturn(m)
		addAPI(CanonicalAPI{
			Module:        cleanModule,
			Name:          name,
			NormalizedKey: fullName,
			FullName:      fullName,
			Kind:          "method",
			RawSignature:  CleanSignature(m.TextRaw),
			Desc:          m.Desc,
			Params:        params,
			Return:        ret,
			Stability:     m.Stability,
			StabilityText: m.StabilityText,
		})
	}

	return apis
}
