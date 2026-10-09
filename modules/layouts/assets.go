package layouts

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"sync"
)

var assetVersions sync.Map

// Asset returns the static asset URL with a content hash query parameter (cache busting).
// Every deploy with a changed file produces a new URL, so browsers fetch the new version.
func Asset(path string) string {
	if v, ok := assetVersions.Load(path); ok {
		return v.(string)
	}

	content, err := os.ReadFile(strings.TrimPrefix(path, "/"))
	if err != nil {
		return path
	}

	hash := sha256.Sum256(content)
	versioned := path + "?v=" + hex.EncodeToString(hash[:])[:12]
	assetVersions.Store(path, versioned)

	return versioned
}
