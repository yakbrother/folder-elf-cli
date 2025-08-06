package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"archive/zip"

	"github.com/fatih/color"
)

// FileOrganizer handles organizing files into categorized folders
type FileOrganizer struct {
	Scanner      *Scanner
	DryRun      bool
	CategoryMap  map[string]string // Maps category names to folder names
	BasePath     string           // Base path where organized folders will be created
}

// NewFileOrganizer creates a new FileOrganizer instance
func NewFileOrganizer(scanner *Scanner, dryRun bool, basePath string) *FileOrganizer {
	// Default category to folder mapping
	categoryMap := map[string]string{
		"Images":       "Images",
		"Documents":    "Documents",
		"Videos":       "Videos",
		"Music":        "Music",
		"Applications": "Applications",
		"Archives":     "Archives",
		"Disk Images":  "Disk Images",
		"Other":        "Other",
	}

	return &FileOrganizer{
		Scanner:     scanner,
		DryRun:     dryRun,
		CategoryMap: categoryMap,
		BasePath:    basePath,
	}
}

// OrganizeFiles organizes all files into their respective category folders
func (fo *FileOrganizer) OrganizeFiles() error {
	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("📁 Starting file organization...")
	fmt.Println()

	totalMoved := 0
	totalSkipped := 0

	// Process each category
	for category, files := range fo.Scanner.Categories {
		folderName, exists := fo.CategoryMap[category]
		if !exists {
			folderName = "Other"
		}

		// Create category folder if it doesn't exist
		categoryPath := filepath.Join(fo.BasePath, folderName)
		if !fo.DryRun {
			err := os.MkdirAll(categoryPath, 0755)
			if err != nil {
				warningColor.Printf("⚠️  Failed to create folder %s: %v\n", folderName, err)
				continue
			}
		}

		infoColor.Printf("📂 Processing %s (%d files)...\n", category, len(files))

		// Move each file to its category folder
		for _, file := range files {
			// Skip duplicate files (they might be removed)
			if file.IsDuplicate {
				continue
			}

			// Skip files that are already in the correct folder
			if filepath.Dir(file.Path) == categoryPath {
				totalSkipped++
				continue
			}

			destPath := filepath.Join(categoryPath, file.Name)

			// Check if destination file already exists
			if _, err := os.Stat(destPath); err == nil {
				warningColor.Printf("⚠️  File already exists at destination: %s\n", destPath)
				totalSkipped++
				continue
			}

			if fo.DryRun {
				fmt.Printf("   📁 Would move: %s -> %s\n", file.Name, folderName)
			} else {
				fmt.Printf("   📁 Moving: %s\n", file.Name)
				err := os.Rename(file.Path, destPath)
				if err != nil {
					// If rename fails (cross-device?), try copy + delete
					err = fo.copyAndDelete(file.Path, destPath)
					if err != nil {
						warningColor.Printf("   ⚠️  Failed to move %s: %v\n", file.Name, err)
						totalSkipped++
						continue
					}
				}
			}
			totalMoved++
		}
		fmt.Println()
	}

	if totalMoved > 0 {
		successColor.Printf("✅ Moved %d files to organized folders!\n", totalMoved)
	}
	if totalSkipped > 0 {
		fmt.Printf("ℹ️  Skipped %d files (already in place or conflicts)\n", totalSkipped)
	}

	return nil
}

// OrganizeByDate organizes files into date-based folders (YYYY-MM format)
func (fo *FileOrganizer) OrganizeByDate() error {
	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("📅 Starting date-based organization...")
	fmt.Println()

	totalMoved := 0
	totalSkipped := 0

	// Group files by date
	dateGroups := make(map[string][]FileInfo)
	for _, file := range fo.Scanner.Files {
		if file.IsDuplicate {
			continue
		}

		// Get year-month from modification date
		dateKey := file.LastModified.Format("2006-01")
		dateGroups[dateKey] = append(dateGroups[dateKey], file)
	}

	// Process each date group
	for dateKey, files := range dateGroups {
		// Create date folder
		datePath := filepath.Join(fo.BasePath, dateKey)
		if !fo.DryRun {
			err := os.MkdirAll(datePath, 0755)
			if err != nil {
				warningColor.Printf("⚠️  Failed to create folder %s: %v\n", dateKey, err)
				continue
			}
		}

		infoColor.Printf("📅 Processing %s (%d files)...\n", dateKey, len(files))

		// Move each file to its date folder
		for _, file := range files {
			// Skip files that are already in the correct folder
			if filepath.Dir(file.Path) == datePath {
				totalSkipped++
				continue
			}

			destPath := filepath.Join(datePath, file.Name)

			// Check if destination file already exists
			if _, err := os.Stat(destPath); err == nil {
				warningColor.Printf("⚠️  File already exists at destination: %s\n", destPath)
				totalSkipped++
				continue
			}

			if fo.DryRun {
				fmt.Printf("   📁 Would move: %s -> %s\n", file.Name, dateKey)
			} else {
				fmt.Printf("   📁 Moving: %s\n", file.Name)
				err := os.Rename(file.Path, destPath)
				if err != nil {
					// If rename fails (cross-device?), try copy + delete
					err = fo.copyAndDelete(file.Path, destPath)
					if err != nil {
						warningColor.Printf("   ⚠️  Failed to move %s: %v\n", file.Name, err)
						totalSkipped++
						continue
					}
				}
			}
			totalMoved++
		}
		fmt.Println()
	}

	if totalMoved > 0 {
		successColor.Printf("✅ Moved %d files to date-based folders!\n", totalMoved)
	}
	if totalSkipped > 0 {
		fmt.Printf("ℹ️  Skipped %d files (already in place or conflicts)\n", totalSkipped)
	}

	return nil
}

// OrganizeBySize organizes files into size-based folders
func (fo *FileOrganizer) OrganizeBySize() error {
	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("📏 Starting size-based organization...")
	fmt.Println()

	// Define size categories
	sizeCategories := []struct {
		name  string
		min   int64
		max   int64
	}{
		{"Tiny", 0, 1024 * 1024},         // < 1MB
		{"Small", 1024 * 1024, 10 * 1024 * 1024},    // 1MB - 10MB
		{"Medium", 10 * 1024 * 1024, 100 * 1024 * 1024}, // 10MB - 100MB
		{"Large", 100 * 1024 * 1024, 1024 * 1024 * 1024}, // 100MB - 1GB
		{"Huge", 1024 * 1024 * 1024, -1}, // > 1GB
	}

	totalMoved := 0
	totalSkipped := 0

	// Process each size category
	for _, sizeCat := range sizeCategories {
		var filesToMove []FileInfo
		
		for _, file := range fo.Scanner.Files {
			if file.IsDuplicate {
				continue
			}

			if (sizeCat.min == -1 || file.Size >= sizeCat.min) && 
			   (sizeCat.max == -1 || file.Size < sizeCat.max) {
				filesToMove = append(filesToMove, file)
			}
		}

		if len(filesToMove) == 0 {
			continue
		}

		// Create size folder
		sizePath := filepath.Join(fo.BasePath, sizeCat.name)
		if !fo.DryRun {
			err := os.MkdirAll(sizePath, 0755)
			if err != nil {
				warningColor.Printf("⚠️  Failed to create folder %s: %v\n", sizeCat.name, err)
				continue
			}
		}

		infoColor.Printf("📏 Processing %s files (%d files)...\n", sizeCat.name, len(filesToMove))

		// Move each file to its size folder
		for _, file := range filesToMove {
			// Skip files that are already in the correct folder
			if filepath.Dir(file.Path) == sizePath {
				totalSkipped++
				continue
			}

			destPath := filepath.Join(sizePath, file.Name)

			// Check if destination file already exists
			if _, err := os.Stat(destPath); err == nil {
				warningColor.Printf("⚠️  File already exists at destination: %s\n", destPath)
				totalSkipped++
				continue
			}

			if fo.DryRun {
				fmt.Printf("   📁 Would move: %s -> %s\n", file.Name, sizeCat.name)
			} else {
				fmt.Printf("   📁 Moving: %s\n", file.Name)
				err := os.Rename(file.Path, destPath)
				if err != nil {
					// If rename fails (cross-device?), try copy + delete
					err = fo.copyAndDelete(file.Path, destPath)
					if err != nil {
						warningColor.Printf("   ⚠️  Failed to move %s: %v\n", file.Name, err)
						totalSkipped++
						continue
					}
				}
			}
			totalMoved++
		}
		fmt.Println()
	}

	if totalMoved > 0 {
		successColor.Printf("✅ Moved %d files to size-based folders!\n", totalMoved)
	}
	if totalSkipped > 0 {
		fmt.Printf("ℹ️  Skipped %d files (already in place or conflicts)\n", totalSkipped)
	}

	return nil
}

// ProcessZipFiles processes zip files and organizes their contents
func (fo *FileOrganizer) ProcessZipFiles() error {
	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("📦 Starting zip file processing...")
	fmt.Println()

	totalProcessed := 0
	totalSkipped := 0

	// Get all zip files
	zipFiles := fo.Scanner.Categories["Archives"]
	if len(zipFiles) == 0 {
		fmt.Println("ℹ️  No zip files found to process.")
		return nil
	}

	for _, zipFile := range zipFiles {
		if zipFile.IsDuplicate {
			continue
		}

		infoColor.Printf("📦 Processing zip file: %s\n", zipFile.Name)

		// Open the zip file
		r, err := zip.OpenReader(zipFile.Path)
		if err != nil {
			warningColor.Printf("⚠️  Failed to open zip file %s: %v\n", zipFile.Name, err)
			totalSkipped++
			continue
		}
		defer r.Close()

		// Analyze zip contents to determine the best category
		category := fo.analyzeZipContents(&r.Reader)
		infoColor.Printf("   📂 Zip appears to contain: %s\n", category)

		// Create category folder if it doesn't exist
		folderName, exists := fo.CategoryMap[category]
		if !exists {
			folderName = "Other"
		}

		categoryPath := filepath.Join(fo.BasePath, folderName)
		if !fo.DryRun {
			err := os.MkdirAll(categoryPath, 0755)
			if err != nil {
				warningColor.Printf("   ⚠️  Failed to create folder %s: %v\n", folderName, err)
				totalSkipped++
				continue
			}
		}

		// Move the zip file to the appropriate category
		destPath := filepath.Join(categoryPath, zipFile.Name)

		if fo.DryRun {
			fmt.Printf("   📁 Would move: %s -> %s\n", zipFile.Name, folderName)
		} else {
			fmt.Printf("   📁 Moving: %s\n", zipFile.Name)
			err := os.Rename(zipFile.Path, destPath)
			if err != nil {
				// If rename fails (cross-device?), try copy + delete
				err = fo.copyAndDelete(zipFile.Path, destPath)
				if err != nil {
					warningColor.Printf("   ⚠️  Failed to move %s: %v\n", zipFile.Name, err)
					totalSkipped++
					continue
				}
			}
		}
		totalProcessed++
		fmt.Println()
	}

	if totalProcessed > 0 {
		successColor.Printf("✅ Processed %d zip files!\n", totalProcessed)
	}
	if totalSkipped > 0 {
		fmt.Printf("ℹ️  Skipped %d zip files\n", totalSkipped)
	}

	return nil
}

// analyzeZipContents analyzes the contents of a zip file to determine its category
func (fo *FileOrganizer) analyzeZipContents(r *zip.Reader) string {
	imageCount := 0
	documentCount := 0
	videoCount := 0
	audioCount := 0
	applicationCount := 0
	fontCount := 0
	codeCount := 0

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg", ".webp":
			imageCount++
		case ".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt", ".xls", ".xlsx", ".ppt", ".pptx":
			documentCount++
		case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
			videoCount++
		case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma":
			audioCount++
		case ".exe", ".msi", ".dmg", ".pkg", ".app", ".deb", ".rpm":
			applicationCount++
		case ".ttf", ".otf", ".woff", ".woff2", ".eot":
			fontCount++
		case ".js", ".py", ".java", ".cpp", ".c", ".cs", ".php", ".rb", ".go", ".rs", ".swift", ".kt", ".html", ".css", ".scss", ".sql", ".sh", ".json", ".xml", ".yaml", ".yml":
			codeCount++
		}
	}

	// Determine the dominant category
	maxCount := 0
	dominantCategory := "Other"

	if imageCount > maxCount {
		maxCount = imageCount
		dominantCategory = "Images"
	}
	if documentCount > maxCount {
		maxCount = documentCount
		dominantCategory = "Documents"
	}
	if videoCount > maxCount {
		maxCount = videoCount
		dominantCategory = "Videos"
	}
	if audioCount > maxCount {
		maxCount = audioCount
		dominantCategory = "Music"
	}
	if applicationCount > maxCount {
		maxCount = applicationCount
		dominantCategory = "Applications"
	}
	if fontCount > maxCount {
		maxCount = fontCount
		dominantCategory = "Other" // Fonts go to Other or could have their own category
	}
	if codeCount > maxCount {
		maxCount = codeCount
		dominantCategory = "Other" // Code goes to Other or could have its own category
	}

	return dominantCategory
}

// copyAndDelete copies a file to destination and then deletes the original
func (fo *FileOrganizer) copyAndDelete(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Delete source file
	return os.Remove(src)
}