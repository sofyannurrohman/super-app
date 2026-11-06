package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// SaveUploadedFile uploads a multipart file to the destination path safely without
// attempting to chmod the directory if it already exists, preventing "operation not permitted"
// errors when directories are owned by different users (e.g. root in docker mounts).
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Ensure the parent directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the contents
	_, err = io.Copy(out, src)
	return err
}
