package model

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
)

// NormalizePath ensures paths are relative to the repository root and use forward slashes.
func NormalizePath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	return p
}

// FingerprintStable computes a deterministic fingerprint for a finding.
// stableKey should be something like "aws_access_key" or "pem_private_key", never the matched secret itself.
func FingerprintStable(ruleID, path string, line int, stableKey string) string {
	path = NormalizePath(path)
	base := ruleID + "|" + path + "|"
	if line > 0 {
		base += strconv.Itoa(line) + "|"
	} else {
		base += "0|"
	}
	base += stableKey
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}
