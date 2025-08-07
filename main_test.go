package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid home directory path",
			path:    os.Getenv("HOME") + "/Downloads",
			wantErr: false,
		},
		{
			name:    "valid temp directory path",
			path:    os.TempDir(),
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with traversal",
			path:    "/tmp/../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with double dots",
			path:    "/home/user/../../etc",
			wantErr: true,
		},
		{
			name:    "system directory outside allowed",
			path:    "/etc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultDownloadsPath(t *testing.T) {
	path, err := getDefaultDownloadsPath()
	if err != nil {
		t.Errorf("getDefaultDownloadsPath() error = %v", err)
		return
	}

	if path == "" {
		t.Error("getDefaultDownloadsPath() returned empty path")
	}

	// Check if the path exists or can be created
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Logf("Downloads directory does not exist: %s", path)
	} else if err != nil {
		t.Errorf("Error checking downloads path: %v", err)
	}
}

func TestFileInfo(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test FileInfo creation
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	fileInfo := FileInfo{
		Path:         tmpFile,
		Name:         info.Name(),
		Size:         info.Size(),
		Extension:    ".txt",
		Category:     "Documents",
		Hash:         "",
		LastModified: info.ModTime(),
		IsDuplicate:  false,
		IsZip:        false,
	}

	if fileInfo.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", fileInfo.Name)
	}

	if fileInfo.Size != int64(len("test content")) {
		t.Errorf("Expected size %d, got %d", len("test content"), fileInfo.Size)
	}

	if fileInfo.Category != "Documents" {
		t.Errorf("Expected category 'Documents', got '%s'", fileInfo.Category)
	}
} 