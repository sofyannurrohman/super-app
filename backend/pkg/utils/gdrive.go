package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GDriveService is a wrapper around the Google Drive API
type GDriveService struct {
	srv      *drive.Service
	folderID string
}

// NewGDriveService initializes the Google Drive API client using a service account or OAuth client
func NewGDriveService(credentialsFile string, folderID string) (*GDriveService, error) {
	ctx := context.Background()

	var srv *drive.Service
	b, err := os.ReadFile(credentialsFile)
	if err == nil {
		// Check if it's an OAuth client file (contains "installed" or "web")
		var secretMap map[string]interface{}
		if json.Unmarshal(b, &secretMap) == nil && (secretMap["installed"] != nil || secretMap["web"] != nil) {
			// This is an OAuth Client Secret. We must load token.json
			config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
			if err != nil {
				return nil, fmt.Errorf("failed to parse client secret: %v", err)
			}
			tokFile := "token.json"
			tokBytes, err := os.ReadFile(tokFile)
			if err != nil {
				return nil, fmt.Errorf("token.json tidak ditemukan. Silakan jalankan 'go run generate_token.go %s' terlebih dahulu", credentialsFile)
			}
			var tok oauth2.Token
			if err := json.Unmarshal(tokBytes, &tok); err != nil {
				return nil, fmt.Errorf("failed to parse token.json: %v", err)
			}
			client := config.Client(ctx, &tok)
			srv, err = drive.NewService(ctx, option.WithHTTPClient(client))
		} else {
			// Service Account or ADC
			srv, err = drive.NewService(ctx, option.WithCredentialsFile(credentialsFile))
		}
	} else {
		// Fallback to ADC if file doesn't exist
		srv, err = drive.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %v", err)
	}

	return &GDriveService{
		srv:      srv,
		folderID: folderID,
	}, nil
}

// UploadFile uploads a file to the configured Google Drive folder and returns the public link
func (g *GDriveService) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Ensure file is read from the beginning
	_, _ = file.Seek(0, 0)

	// Create the file metadata
	f := &drive.File{
		Name:    header.Filename,
		Parents: []string{g.folderID},
	}

	// Upload the file (SupportsAllDrives allows uploading to Shared Drives)
	res, err := g.srv.Files.Create(f).Media(file).SupportsAllDrives(true).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Make the file publicly accessible
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = g.srv.Permissions.Create(res.Id, permission).SupportsAllDrives(true).Do()
	if err != nil {
		log.Printf("Warning: could not set public permission on file %s: %v", res.Id, err)
		// We still return the link even if permission fails (it might be managed at folder level)
	}

	// Return the sharing link
	return fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", res.Id), nil
}

// UploadFileWithName uploads a file to the configured Google Drive folder using a specific
// filename, preserving the original name for traceability and cross-validation purposes.
// The customName parameter allows callers to inject a name like "[bannerID]_originalname.pdf".
func (g *GDriveService) UploadFileWithName(file multipart.File, customName string) (string, error) {
	// Ensure file is read from the beginning
	_, _ = file.Seek(0, 0)

	f := &drive.File{
		Name:    customName,
		Parents: []string{g.folderID},
	}

	res, err := g.srv.Files.Create(f).Media(file).SupportsAllDrives(true).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload file to GDrive: %v", err)
	}

	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = g.srv.Permissions.Create(res.Id, permission).SupportsAllDrives(true).Do()
	if err != nil {
		log.Printf("Warning: could not set public permission on GDrive file %s: %v", res.Id, err)
	}

	return fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", res.Id), nil
}

// UploadLocalFile uploads a file from the server local filesystem directly to Google Drive.
// This is used for chunked uploads or large video files where chunks are assembled on disk first.
func (g *GDriveService) UploadLocalFile(localFilePath string, customName string) (string, error) {
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open local file for GDrive upload: %w", err)
	}
	defer file.Close()

	f := &drive.File{
		Name:    customName,
		Parents: []string{g.folderID},
	}

	res, err := g.srv.Files.Create(f).Media(file).SupportsAllDrives(true).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload local file to GDrive: %w", err)
	}

	// Make file readable publicly for in-app video streaming
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = g.srv.Permissions.Create(res.Id, permission).SupportsAllDrives(true).Do()
	if err != nil {
		log.Printf("Warning: could not set public permission on GDrive file %s: %v", res.Id, err)
	}

	return fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", res.Id), nil
}

// ExtractGDriveFileID extracts the file ID from a Google Drive URL
func ExtractGDriveFileID(urlStr string) string {
	// Patterns:
	// https://drive.google.com/file/d/{FILE_ID}/view...
	// https://drive.google.com/file/d/{FILE_ID}/preview
	// https://drive.google.com/uc?id={FILE_ID}&...
	if idx := strings.Index(urlStr, "/file/d/"); idx != -1 {
		sub := urlStr[idx+len("/file/d/"):]
		if slashIdx := strings.Index(sub, "/"); slashIdx != -1 {
			return sub[:slashIdx]
		}
		if qIdx := strings.Index(sub, "?"); qIdx != -1 {
			return sub[:qIdx]
		}
		return sub
	}
	if idx := strings.Index(urlStr, "id="); idx != -1 {
		sub := urlStr[idx+3:]
		if ampIdx := strings.Index(sub, "&"); ampIdx != -1 {
			return sub[:ampIdx]
		}
		return sub
	}
	return ""
}

// DeleteFileByURL attempts to delete a file from Google Drive given its sharing URL.
func (g *GDriveService) DeleteFileByURL(urlStr string) error {
	fileID := ExtractGDriveFileID(urlStr)
	if fileID == "" {
		return nil
	}
	err := g.srv.Files.Delete(fileID).SupportsAllDrives(true).Do()
	if err != nil {
		log.Printf("[GDrive] Warning: gagal menghapus file GDrive %s: %v", fileID, err)
		return err
	}
	log.Printf("[GDrive] Berhasil menghapus file GDrive: %s", fileID)
	return nil
}
