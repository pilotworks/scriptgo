package pkgmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func linkBins(nodeModules string, nodes map[string]*packageNode, rootEdges map[string]string) error {
	binDir := filepath.Join(nodeModules, ".bin")
	for _, identity := range rootEdges {
		node := nodes[identity]
		if node == nil {
			continue
		}
		for name, target := range node.Manifest.Bin {
			if name == "" {
				name = node.Name
			}
			if filepath.Base(name) != name || filepath.IsAbs(target) || strings.HasPrefix(filepath.Clean(filepath.FromSlash(target)), ".."+string(filepath.Separator)) {
				return fmt.Errorf("invalid bin declaration for %s", node.Name)
			}
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				return err
			}
			if err := linkBin(binDir, name, node.Name, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func linkBin(binDir, name, packageName, target string) error {
	linkTarget := filepath.Join("..", filepath.FromSlash(packageName), filepath.FromSlash(target))
	if runtime.GOOS != "windows" {
		if err := os.Symlink(linkTarget, filepath.Join(binDir, name)); err != nil {
			return fmt.Errorf("link bin %s: %w", name, err)
		}
		return nil
	}
	cmdTarget := strings.ReplaceAll(linkTarget, "/", "\\")
	relativeTarget := strings.TrimPrefix(cmdTarget, `..\`)
	cmd := fmt.Sprintf("@echo off\r\n\"%%~dp0\\%s\" %%*\r\n", relativeTarget)
	if err := os.WriteFile(filepath.Join(binDir, name+".cmd"), []byte(cmd), 0o755); err != nil {
		return fmt.Errorf("write bin shim %s: %w", name, err)
	}
	powershellTarget := strings.ReplaceAll(relativeTarget, `\`, "/")
	powershell := fmt.Sprintf("& (Join-Path $PSScriptRoot '%s') $args\r\n", powershellTarget)
	if err := os.WriteFile(filepath.Join(binDir, name+".ps1"), []byte(powershell), 0o755); err != nil {
		return fmt.Errorf("write PowerShell bin shim %s: %w", name, err)
	}
	return nil
}
