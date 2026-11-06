package utils

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
)

// CompressAndSaveImage takes an uploaded file, resizes it to a max width (e.g., 800px) while maintaining aspect ratio,
// compresses it as JPEG, and saves it to the specified destination path.
func CompressAndSaveImage(file *multipart.FileHeader, destPath string, maxWidth uint) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	var img image.Image
	ext := filepath.Ext(file.Filename)
	switch ext {
	case ".jpg", ".jpeg", ".JPG", ".JPEG":
		img, err = jpeg.Decode(src)
	case ".png", ".PNG":
		img, err = png.Decode(src)
	default:
		return errors.New("unsupported image format")
	}

	if err != nil {
		return err
	}

	// Resize to max width, 0 height means maintaining aspect ratio
	m := resize.Resize(maxWidth, 0, img, resize.Lanczos3)

	// ensure directory exists
	os.MkdirAll(filepath.Dir(destPath), os.ModePerm)

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode as JPEG with quality 75 (good balance between size and quality)
	var opt jpeg.Options
	opt.Quality = 75
	return jpeg.Encode(out, m, &opt)
}
