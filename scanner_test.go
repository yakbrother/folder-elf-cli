package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []struct {
		name     string
		content  string
		category string
	}{
		{"image.jpg", "fake image data", "Images"},
		{"document.pdf", "fake pdf data", "Documents"},
		{"video.mp4", "fake video data", "Videos"},
		{"music.mp3", "fake music data", "Music"},
		{"archive.zip", "fake zip data", "Archives"},
		{"app.exe", "fake exe data", "Applications"},
		{"disk.iso", "fake iso data", "Disk Images"},
		{"unknown.xyz", "fake data", "Other"},
	}

	// Create the test files
	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	// Create a scanner and scan the directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Test that all files were found
	if len(scanner.Files) != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), len(scanner.Files))
	}

	// Test categorization
	for _, tf := range testFiles {
		found := false
		for _, file := range scanner.Files {
			if file.Name == tf.name {
				if file.Category != tf.category {
					t.Errorf("File %s: expected category '%s', got '%s'", tf.name, tf.category, file.Category)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File %s not found in scan results", tf.name)
		}
	}

	// Test categories map
	for category, files := range scanner.Categories {
		if len(files) == 0 {
			t.Errorf("Category '%s' has no files", category)
		}
	}
}

func TestDetermineCategory(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{"image.jpg", ".jpg", "Images"},
		{"document.pdf", ".pdf", "Documents"},
		{"video.mp4", ".mp4", "Videos"},
		{"music.mp3", ".mp3", "Music"},
		{"archive.zip", ".zip", "Archives"},
		{"app.exe", ".exe", "Applications"},
		{"disk.iso", ".iso", "Disk Images"},
		{"unknown.xyz", ".xyz", "Other"},
		{"installer.exe", ".exe", "Applications"},
		{"setup.exe", ".exe", "Applications"},
		{"manual.pdf", ".pdf", "Documents"},
		{"guide.txt", ".txt", "Documents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.determineCategory(tt.ext, tt.name)
			if result != tt.expected {
				t.Errorf("determineCategory(%s, %s) = %s, want %s", tt.ext, tt.name, result, tt.expected)
			}
		})
	}
}

func TestCalculateFileHash(t *testing.T) {
	scanner := NewScanner()

	// Create a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate hash
	hash, err := scanner.calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash() error = %v", err)
	}

	if hash == "" {
		t.Error("calculateFileHash() returned empty hash")
	}

	// Test that same content produces same hash
	hash2, err := scanner.calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash() error on second call: %v", err)
	}

	if hash != hash2 {
		t.Errorf("Hashes don't match: %s != %s", hash, hash2)
	}
}

func TestFindDuplicates(t *testing.T) {
	scanner := NewScanner()

	// Create test files with same content (should have same hash)
	tmpDir := t.TempDir()
	testContent := "duplicate content"
	
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	
	for _, filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Scan directory
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Check that duplicates were found
	if len(scanner.Duplicates) == 0 {
		t.Error("Expected duplicates to be found, but none were detected")
	}

	// Check that files are marked as duplicates
	duplicateCount := 0
	for _, file := range scanner.Files {
		if file.IsDuplicate {
			duplicateCount++
		}
	}

	if duplicateCount == 0 {
		t.Error("Expected files to be marked as duplicates")
	}
}

func TestCheckFilePermissions(t *testing.T) {
	scanner := NewScanner()

	// Create a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that we can read the file
	err = scanner.checkFilePermissions(testFile)
	if err != nil {
		t.Errorf("checkFilePermissions() error = %v", err)
	}

	// Test with non-existent file
	err = scanner.checkFilePermissions("/non/existent/file")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
} 