package llvm

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type locationKey struct {
	line  int
	col   int
	scope int
}

type locationEntry struct {
	id    int
	line  int
	col   int
	scope int
}

type debugInfo struct {
	files      map[string]int
	functions  map[string]int
	locations  map[locationKey]int
	locList    []locationEntry
	nextID     int
	moduleFile int
}

func newDebugInfo(module ir.Module) *debugInfo {
	paths := map[string]bool{}
	for _, function := range module.Functions {
		if function.Span.Path != "" {
			paths[function.Span.Path] = true
		}
	}
	if module.SourcePath != "" {
		paths[module.SourcePath] = true
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	result := &debugInfo{
		files:     make(map[string]int),
		functions: make(map[string]int),
		locations: make(map[locationKey]int),
		nextID:    4 + len(ordered),
	}
	for index, path := range ordered {
		result.files[path] = index + 4
	}
	if module.SourcePath != "" {
		result.moduleFile = result.files[module.SourcePath]
	}
	for _, function := range module.Functions {
		result.functions[function.Name] = result.nextID
		result.nextID++
	}
	return result
}

func (info *debugInfo) location(span ir.SourceSpan, funcName string, module ir.Module) string {
	if info == nil {
		return ""
	}
	scope, ok := info.functions[funcName]
	if !ok {
		return ""
	}
	path := span.Path
	if path == "" {
		path = module.SourcePath
	}
	source := module.SourceFiles[path]
	line := sourceLine(source, span.Offset)
	col := sourceColumn(source, span.Offset)
	key := locationKey{line: line, col: col, scope: scope}
	id, exists := info.locations[key]
	if !exists {
		id = info.nextID
		info.nextID++
		info.locations[key] = id
		info.locList = append(info.locList, locationEntry{id: id, line: line, col: col, scope: scope})
	}
	return fmt.Sprintf("!dbg !%d", id)
}

func (info *debugInfo) functionMetadata(function ir.Function, module ir.Module) string {
	span := function.Span
	if span.Path == "" && len(function.Body) > 0 {
		span = function.Body[0].Span
	}
	file := info.files[span.Path]
	if file == 0 {
		file = info.moduleFile
	}
	line := sourceLine(module.SourceFiles[span.Path], span.Offset)
	return fmt.Sprintf("!%d = distinct !DISubprogram(name: %q, linkageName: %q, scope: !%d, file: !%d, line: %d, type: !3, scopeLine: %d, spFlags: DISPFlagDefinition, unit: !0)", info.functions[function.Name], function.Name, function.Name, file, file, line, line)
}

func (info *debugInfo) metadata(module ir.Module, compilerVersion string) string {
	paths := make([]string, 0, len(info.files))
	for path := range info.files {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return info.files[paths[i]] < info.files[paths[j]] })
	var text strings.Builder
	text.WriteString("\n!llvm.dbg.cu = !{!0}\n")
	text.WriteString(fmt.Sprintf("!0 = distinct !DICompileUnit(language: DW_LANG_C, file: !%d, producer: %q, isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)\n", info.moduleFile, "scriptgo "+compilerVersion))
	for _, path := range paths {
		text.WriteString(fmt.Sprintf("!%d = !DIFile(filename: %q, directory: \".\")\n", info.files[path], filepath.Base(path)))
	}
	text.WriteString("!3 = !DISubroutineType(types: !{})\n")
	for _, functionPath := range module.Functions {
		text.WriteString(info.functionMetadata(functionPath, module) + "\n")
	}
	for _, loc := range info.locList {
		text.WriteString(fmt.Sprintf("!%d = !DILocation(line: %d, column: %d, scope: !%d)\n", loc.id, loc.line, loc.col, loc.scope))
	}
	text.WriteString("!llvm.module.flags = !{!1}\n!1 = !{i32 2, !\"Debug Info Version\", i32 3}\n")
	return text.String()
}

func sourceLine(source string, offset int) int {
	if offset < 0 {
		return 1
	}
	if offset > len(source) {
		offset = len(source)
	}
	line := 1
	for index := 0; index < offset; index++ {
		if source[index] == '\n' {
			line++
		}
	}
	return line
}

func sourceColumn(source string, offset int) int {
	if offset < 0 {
		return 1
	}
	if offset > len(source) {
		offset = len(source)
	}
	lastNewline := strings.LastIndex(source[:offset], "\n")
	if lastNewline == -1 {
		return offset + 1
	}
	return offset - lastNewline
}
