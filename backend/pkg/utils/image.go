package utils

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// CompressImage decodes an image from a byte slice, resizes it to a maximum width
// while preserving aspect ratio, and re-encodes it as a JPEG with the specified quality.
func CompressImage(data []byte, maxWidth int, quality int) ([]byte, error) {
	reader := bytes.NewReader(data)

	// Decode the image
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, err
	}

	// Resize the image if its width is greater than maxWidth
	if img.Bounds().Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	// Encode it back to a byte slice as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SaveUploadedImageCompressed reads a multipart uploaded image, applies auto-orientation,
// resizes it to maxWidth if necessary, and compresses/saves it directly as an optimized JPEG.
// It is 100% Pure Go without any CGo/C compiler dependencies (ideal for Docker Alpine & Linux).
// Returns the relative storage path (e.g., "/uploads/banners/installed/xxx.jpg").
func SaveUploadedImageCompressed(header *multipart.FileHeader, targetDir string, maxWidth int, quality int) (string, error) {
	src, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	// Read image bytes
	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file bytes: %w", err)
	}

	// Check if this is a supported image by trying to decode it
	reader := bytes.NewReader(data)
	img, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		// If decoding fails (e.g. PDF or non-raster format), fallback to standard file saving
		rawFilename := fmt.Sprintf("%s-%s", uuid.New().String(), filepath.Base(header.Filename))
		rawPath := filepath.Join(targetDir, rawFilename)
		if saveErr := SaveUploadedFile(header, rawPath); saveErr != nil {
			return "", saveErr
		}
		cleanDir := strings.TrimPrefix(filepath.ToSlash(targetDir), "./")
		if !strings.HasPrefix(cleanDir, "/") {
			cleanDir = "/" + cleanDir
		}
		return filepath.ToSlash(filepath.Join(cleanDir, rawFilename)), nil
	}

	// Resize image if width exceeds maxWidth
	if maxWidth > 0 && img.Bounds().Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	if quality <= 0 || quality > 100 {
		quality = 80
	}

	// Save compressed image as JPG
	jpgFilename := fmt.Sprintf("%s.jpg", uuid.New().String())
	jpgPath := filepath.Join(targetDir, jpgFilename)

	if err := imaging.Save(img, jpgPath, imaging.JPEGQuality(quality)); err != nil {
		return "", fmt.Errorf("failed to save compressed image: %w", err)
	}

	cleanDir := strings.TrimPrefix(filepath.ToSlash(targetDir), "./")
	if !strings.HasPrefix(cleanDir, "/") {
		cleanDir = "/" + cleanDir
	}

	return filepath.ToSlash(filepath.Join(cleanDir, jpgFilename)), nil
}

// SaveUploadedImageAsWebP is an alias to SaveUploadedImageCompressed for backward compatibility.
func SaveUploadedImageAsWebP(header *multipart.FileHeader, targetDir string, maxWidth int, quality float32) (string, error) {
	return SaveUploadedImageCompressed(header, targetDir, maxWidth, int(quality))
}
