package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo holds information about a file
type FileInfo struct {
	Path         string
	Name         string
	Size         int64
	Extension    string
	Category     string
	Hash         string
	LastModified time.Time
	IsDuplicate  bool
	IsZip        bool
}

// Scanner handles scanning the downloads folder
type Scanner struct {
	Files      []FileInfo
	Duplicates map[string][]FileInfo // Map of hash to files with that hash
	Categories map[string][]FileInfo // Map of category to files in that category
}

// NewScanner creates a new Scanner instance
func NewScanner() *Scanner {
	return &Scanner{
		Files:      make([]FileInfo, 0),
		Duplicates: make(map[string][]FileInfo),
		Categories: make(map[string][]FileInfo),
	}
}

// checkFilePermissions checks if we have read permissions for a file
func (s *Scanner) checkFilePermissions(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %v", filePath, err)
	}
	file.Close()
	return nil
}

// ScanDirectory scans a directory and collects file information
func (s *Scanner) ScanDirectory(dirPath string) error {
	fmt.Printf("🔍 Scanning directory: %s\n", dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories (like .DS_Store on macOS)
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			
			// Skip macOS .app bundle contents
			if strings.HasSuffix(info.Name(), ".app") {
				return filepath.SkipDir
			}
			
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		// Skip files inside .app bundles
		if strings.Contains(path, ".app/Contents/") {
			return nil
		}

		// Check file permissions before processing
		if err := s.checkFilePermissions(path); err != nil {
			fmt.Printf("⚠️  Skipping file due to permission error: %s - %v\n", path, err)
			return nil // Continue scanning other files
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == "" {
			ext = "no_extension"
		}

		// Determine category
		category := s.determineCategory(ext, info.Name())

		// Calculate file hash for duplicate detection
		hash, err := s.calculateFileHash(path)
		if err != nil {
			fmt.Printf("⚠️  Could not calculate hash for %s: %v\n", path, err)
			// Continue without hash rather than failing completely
			hash = ""
		}

		// Create file info
		fileInfo := FileInfo{
			Path:         path,
			Name:         info.Name(),
			Size:         info.Size(),
			Extension:    ext,
			Category:     category,
			Hash:         hash,
			LastModified: info.ModTime(),
			IsZip:        ext == ".zip",
		}

		s.Files = append(s.Files, fileInfo)

		// Add to categories map
		s.Categories[category] = append(s.Categories[category], fileInfo)

		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning directory: %v", err)
	}

	// Find duplicates after scanning all files
	s.findDuplicates()

	fmt.Printf("✅ Found %d files\n", len(s.Files))
	return nil
}

// determineCategory determines the category of a file based on its extension and name
func (s *Scanner) determineCategory(ext, name string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg", ".webp":
		return "Images"
	case ".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt", ".xls", ".xlsx", ".ppt", ".pptx":
		return "Documents"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
		return "Videos"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma":
		return "Music"
	case ".pkg", ".exe", ".msi", ".deb", ".rpm", ".app":
		return "Applications"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2":
		return "Archives"
	case ".iso", ".dmg":
		return "Disk Images"
	default:
		// Try to determine from name patterns
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "install") || strings.Contains(lowerName, "setup") {
			return "Applications"
		}
		if strings.Contains(lowerName, "manual") || strings.Contains(lowerName, "guide") {
			return "Documents"
		}
		return "Other"
	}
}

// calculateFileHash calculates the MD5 hash of a file
func (s *Scanner) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("⚠️  Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	hash := md5.New()
	// Use a buffer to limit memory usage for large files
	buf := make([]byte, 32*1024) // 32KB buffer
	if _, err := io.CopyBuffer(hash, file, buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// findDuplicates finds duplicate files based on their hash
func (s *Scanner) findDuplicates() {
	fmt.Println("🔍 Checking for duplicates...")

	hashMap := make(map[string][]FileInfo)

	// Group files by hash
	for _, file := range s.Files {
		if file.Hash != "" {
			hashMap[file.Hash] = append(hashMap[file.Hash], file)
		}
	}

	// Find duplicates (files with same hash)
	for hash, files := range hashMap {
		if len(files) > 1 {
			s.Duplicates[hash] = files
			// Mark files as duplicates
			for i := range files {
				// Create a new reference to the file in the Files slice
				for j := range s.Files {
					if s.Files[j].Path == files[i].Path {
						s.Files[j].IsDuplicate = true
						break
					}
				}
			}
		}
	}

	duplicateCount := 0
	for _, files := range s.Duplicates {
		duplicateCount += len(files)
	}

	if duplicateCount > 0 {
		fmt.Printf("⚠️  Found %d duplicate files\n", duplicateCount)
	} else {
		fmt.Println("✅ No duplicates found")
	}
}

// PrintSummary prints a summary of the scan results
func (s *Scanner) PrintSummary() {
	fmt.Println("\n📊 Scan Summary:")
	fmt.Printf("Total files: %d\n", len(s.Files))

	fmt.Println("\n📂 Files by category:")
	for category, files := range s.Categories {
		fmt.Printf("  %s: %d files\n", category, len(files))
	}

	if len(s.Duplicates) > 0 {
		fmt.Println("\n🔄 Duplicate files:")
		for hash, files := range s.Duplicates {
			fmt.Printf("  Hash: %s\n", hash[:8]+"...")
			for _, file := range files {
				fmt.Printf("    - %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
			}
		}
	}
}