package frontend

import (
	"io/fs"

	"github.com/microsoft/TypeScript/tsc/scriptgo"
)

// ParseSyntax parses one file into normalized syntax without type checking.
// It is intended for declaration scanning (for example API coverage audits),
// not for building a checked program.
func ParseSyntax(fileName, source string) (SyntaxFile, error) {
	return typescriptgo.ParseFileToSyntax(fileName, source)
}

// EnsureStdlib loads the embedded standard library declarations for version
// ("" selects the default).
func EnsureStdlib(version string) error {
	return typescriptgo.EnsureStdlib(version)
}

// StdlibFS exposes the embedded standard library declaration files.
func StdlibFS() fs.FS {
	return typescriptgo.EmbeddedFS()
}
