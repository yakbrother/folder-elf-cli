package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFileOrganizer(t *testing.T) {
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

	// Create scanner and scan directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Create file organizer in dry-run mode
	organizer := NewFileOrganizer(scanner, true, tmpDir)

	// Test file organization
	err = organizer.OrganizeFiles()
	if err != nil {
		t.Errorf("OrganizeFiles() error = %v", err)
	}

	// Verify files still exist in original location (dry-run mode)
	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s was moved in dry-run mode", tf.name)
		}
	}
}

func TestOrganizeByDate(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	
	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
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

	// Create file organizer in dry-run mode
	organizer := NewFileOrganizer(scanner, true, tmpDir)

	// Test date-based organization
	err = organizer.OrganizeByDate()
	if err != nil {
		t.Errorf("OrganizeByDate() error = %v", err)
	}

	// Verify files still exist in original location (dry-run mode)
	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s was moved in dry-run mode", filename)
		}
	}
}

func TestOrganizeBySize(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test files of different sizes
	testFiles := []struct {
		name   string
		size   int
	}{
		{"tiny.txt", 500},      // 500 bytes
		{"small.txt", 5 * 1024 * 1024},    // 5MB
		{"medium.txt", 50 * 1024 * 1024},  // 50MB
		{"large.txt", 500 * 1024 * 1024},  // 500MB
		{"huge.txt", 2 * 1024 * 1024 * 1024}, // 2GB
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		// Create file with specified size
		data := make([]byte, tf.size)
		err := os.WriteFile(filePath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	// Create scanner and scan directory
	scanner := NewScanner()
	err := scanner.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	// Create file organizer in dry-run mode
	organizer := NewFileOrganizer(scanner, true, tmpDir)

	// Test size-based organization
	err = organizer.OrganizeBySize()
	if err != nil {
		t.Errorf("OrganizeBySize() error = %v", err)
	}

	// Verify files still exist in original location (dry-run mode)
	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s was moved in dry-run mode", tf.name)
		}
	}
}

func TestCheckZipBomb(t *testing.T) {
	organizer := NewFileOrganizer(nil, true, "")

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test with normal zip file
	normalZip := filepath.Join(tmpDir, "normal.zip")
	err := createTestZip(normalZip, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})
	if err != nil {
		t.Fatalf("Failed to create normal zip: %v", err)
	}

	err = organizer.checkZipBomb(normalZip)
	if err != nil {
		t.Errorf("checkZipBomb() error with normal zip: %v", err)
	}

	// Test with large zip file (should fail)
	largeZip := filepath.Join(tmpDir, "large.zip")
	err = createLargeTestZip(largeZip)
	if err != nil {
		t.Fatalf("Failed to create large zip: %v", err)
	}

	err = organizer.checkZipBomb(largeZip)
	if err == nil {
		t.Error("Expected error for large zip file")
	}
}

func TestAnalyzeZipContents(t *testing.T) {
	organizer := NewFileOrganizer(nil, true, "")

	// Create a test zip file
	tmpDir := t.TempDir()
	testZip := filepath.Join(tmpDir, "test.zip")
	
	err := createTestZip(testZip, map[string]string{
		"image1.jpg": "fake image data",
		"image2.png": "fake image data",
		"doc1.pdf":   "fake pdf data",
		"doc2.txt":   "fake text data",
	})
	if err != nil {
		t.Fatalf("Failed to create test zip: %v", err)
	}

	// Open the zip file
	r, err := zip.OpenReader(testZip)
	if err != nil {
		t.Fatalf("Failed to open test zip: %v", err)
	}
	defer r.Close()

	// Test zip content analysis
	category := organizer.analyzeZipContents(&r.Reader)
	if category != "Images" {
		t.Errorf("Expected category 'Images', got '%s'", category)
	}
}

func TestOrganizerAtomicMove(t *testing.T) {
	organizer := NewFileOrganizer(nil, true, "")

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
	err = organizer.atomicMove(srcFile, dstFile)
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

func TestOrganizerCopyAndDelete(t *testing.T) {
	organizer := NewFileOrganizer(nil, true, "")

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
	err = organizer.copyAndDelete(srcFile, dstFile)
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

// Helper functions for creating test zip files
func createTestZip(zipPath string, files map[string]string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	return nil
}

func createLargeTestZip(zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Create many small files to simulate a large zip
	for i := 0; i < 20000; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		writer, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte("content"))
		if err != nil {
			return err
		}
	}

	return nil
} 