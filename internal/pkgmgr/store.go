package pkgmgr

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Store is the local content-addressed package blob store. It stores immutable
// bytes only; package linking and manifest projection are separate concerns.
type Store struct {
	Root string
}

func (s Store) blobPath(digest string) string {
	return filepath.Join(s.Root, "v1", "sha512", digest[:2], digest)
}

// Put stores bytes under their SHA-512 digest and returns the stable blob path.
// Existing identical blobs are reused without rewriting them.
func (s Store) Put(data []byte) (string, error) {
	if s.Root == "" {
		return "", fmt.Errorf("content store root must not be empty")
	}
	digest := sha512.Sum512(data)
	hexDigest := hex.EncodeToString(digest[:])
	destination := s.blobPath(hexDigest)
	if info, err := os.Stat(destination); err == nil && info.Size() == int64(len(data)) {
		return destination, nil
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", fmt.Errorf("create content store directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".tmp-")
	if err != nil {
		return "", fmt.Errorf("create content store temporary file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return "", fmt.Errorf("set content store permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return "", fmt.Errorf("write content store blob: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", fmt.Errorf("sync content store blob: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close content store blob: %w", err)
	}
	if err := os.Rename(temporaryName, destination); err != nil {
		if info, statErr := os.Stat(destination); statErr == nil && info.Size() == int64(len(data)) {
			return destination, nil
		}
		return "", fmt.Errorf("publish content store blob: %w", err)
	}
	return destination, nil
}

// Read returns a previously stored immutable blob by its hex SHA-512 digest.
func (s Store) Read(digest string) ([]byte, error) {
	if len(digest) != sha512.Size*2 {
		return nil, fmt.Errorf("invalid SHA-512 digest %q", digest)
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return nil, fmt.Errorf("invalid SHA-512 digest %q: %w", digest, err)
	}
	data, err := os.ReadFile(s.blobPath(digest))
	if err != nil {
		return nil, fmt.Errorf("read content store blob %q: %w", digest, err)
	}
	actual := sha512.Sum512(data)
	if hex.EncodeToString(actual[:]) != digest {
		return nil, fmt.Errorf("content store blob %q failed verification", digest)
	}
	return data, nil
}
