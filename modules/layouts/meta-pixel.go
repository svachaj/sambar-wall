package layouts

import "os"

const defaultMetaPixelID = "2043212523203633"

func RegistrationMetaPixelID() string {
	if metaPixelID := os.Getenv("META_PIXEL_ID"); metaPixelID != "" {
		return metaPixelID
	}

	return defaultMetaPixelID
}
