package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDuplicateHandler(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create duplicate files with same content
	testContent := "duplicate content"
	files := []string{"original.txt", "copy (1).txt", "copy (2).txt"}
	
	for _, filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create scanner and scan directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Create duplicate handler
	handler := NewDuplicateHandler(scanner, true) // Use dry-run mode

	// Test that duplicates were found
	if len(scanner.Duplicates) == 0 {
		t.Error("Expected duplicates to be found")
	}

	// Test pattern-based duplicate removal
	err = handler.RemoveDuplicatesByPattern()
	if err != nil {
		t.Errorf("RemoveDuplicatesByPattern() error = %v", err)
	}
}

func TestIsOriginalFile(t *testing.T) {
	handler := NewDuplicateHandler(nil, true)

	tests := []struct {
		name     string
		expected bool
	}{
		{"original.txt", true},
		{"file.txt", true},
		{"document.pdf", true},
		{"file (1).txt", false},
		{"file (2).txt", false},
		{"file copy.txt", false},
		{"file - copy.txt", false},
		{"file_copy.txt", false},
		{"file-copy.txt", false},
		{"file.copy.txt", false},
		{"file (copy).txt", false},
		{"file - (copy).txt", false},
		{"file_ (copy).txt", false},
		{"file duplicate.txt", false},
		{"file - duplicate.txt", false},
		{"file_duplicate.txt", false},
		{"file-duplicate.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isOriginalFile(tt.name)
			if result != tt.expected {
				t.Errorf("isOriginalFile(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create duplicate files
	testContent := "duplicate content"
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	
	for _, filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create scanner and scan directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Create duplicate handler in dry-run mode
	handler := NewDuplicateHandler(scanner, true)

	// Test duplicate removal
	err = handler.RemoveDuplicates()
	if err != nil {
		t.Errorf("RemoveDuplicates() error = %v", err)
	}

	// Verify files still exist (dry-run mode)
	for _, filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s was deleted in dry-run mode", filename)
		}
	}
}

func TestMoveDuplicatesToFolder(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "duplicates")

	// Create duplicate files
	testContent := "duplicate content"
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	
	for _, filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create scanner and scan directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Create duplicate handler in dry-run mode
	handler := NewDuplicateHandler(scanner, true)

	// Test moving duplicates
	err = handler.MoveDuplicatesToFolder(destDir)
	if err != nil {
		t.Errorf("MoveDuplicatesToFolder() error = %v", err)
	}

	// Verify destination directory doesn't exist (dry-run mode)
	if _, err := os.Stat(destDir); err == nil {
		t.Error("Destination directory was created in dry-run mode")
	}
}

func TestAtomicMove(t *testing.T) {
	handler := NewDuplicateHandler(nil, true)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "destination.txt")

	// Create source file
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test atomic move
	err = handler.atomicMove(srcFile, dstFile)
	if err != nil {
		t.Errorf("atomicMove() error = %v", err)
	}

	// Verify file was moved
	if _, err := os.Stat(srcFile); err == nil {
		t.Error("Source file still exists after move")
	}

	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination file doesn't exist after move")
	}
}

func TestCopyAndDelete(t *testing.T) {
	handler := NewDuplicateHandler(nil, true)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "destination.txt")

	// Create source file
	testContent := "test content"
	err := os.WriteFile(srcFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copy and delete
	err = handler.copyAndDelete(srcFile, dstFile)
	if err != nil {
		t.Errorf("copyAndDelete() error = %v", err)
	}

	// Verify file was copied and deleted
	if _, err := os.Stat(srcFile); err == nil {
		t.Error("Source file still exists after copy and delete")
	}

	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination file doesn't exist after copy")
	}

	// Verify content was copied correctly
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: expected '%s', got '%s'", testContent, string(content))
	}
} 