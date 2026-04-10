package utils

import (
	"bytes"
	"image"
	_ "image/png"

	"github.com/nfnt/resize"
	"github.com/sergeymakinen/go-ico"
)

// ConvertToICO converts PNG data to ICO format (32x32) for Windows compatibility.
func ConvertToICO(pngData []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}

	// Resize to standard tray icon size for better compatibility
	img = resize.Resize(32, 32, img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := ico.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
