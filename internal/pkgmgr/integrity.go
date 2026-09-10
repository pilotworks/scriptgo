package pkgmgr

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

// VerifyIntegrity verifies npm's sha512-<base64> SRI form.
func VerifyIntegrity(data []byte, integrity string) error {
	const prefix = "sha512-"
	if len(integrity) <= len(prefix) || integrity[:len(prefix)] != prefix {
		return fmt.Errorf("unsupported integrity %q", integrity)
	}
	expected, err := base64.StdEncoding.DecodeString(integrity[len(prefix):])
	if err != nil {
		return fmt.Errorf("invalid integrity %q: %w", integrity, err)
	}
	digest := sha512.Sum512(data)
	if string(digest[:]) != string(expected) {
		return fmt.Errorf("integrity mismatch")
	}
	return nil
}
