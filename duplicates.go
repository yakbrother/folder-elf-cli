package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// DuplicateHandler handles the removal of duplicate files
type DuplicateHandler struct {
	Scanner *Scanner
	DryRun  bool
}

// NewDuplicateHandler creates a new DuplicateHandler instance
func NewDuplicateHandler(scanner *Scanner, dryRun bool) *DuplicateHandler {
	return &DuplicateHandler{
		Scanner: scanner,
		DryRun:  dryRun,
	}
}

// atomicMove performs an atomic file move operation
func (dh *DuplicateHandler) atomicMove(src, dst string) error {
	// Try atomic rename first (works on same filesystem)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// If rename fails (cross-device), use copy + delete
	return dh.copyAndDelete(src, dst)
}

// RemoveDuplicates removes duplicate files, keeping the newest version of each
func (dh *DuplicateHandler) RemoveDuplicates() error {
	if len(dh.Scanner.Duplicates) == 0 {
		fmt.Println("✅ No duplicates found to remove!")
		return nil
	}

	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("🔄 Processing duplicate files...")
	
	totalRemoved := 0
	totalSpaceSaved := int64(0)

	for hash, files := range dh.Scanner.Duplicates {
		if len(files) < 2 {
			continue
		}

		// Find the newest file to keep
		newestFile := files[0]
		for _, file := range files {
			if file.LastModified.After(newestFile.LastModified) {
				newestFile = file
			}
		}

		infoColor.Printf("📋 Processing duplicates for hash: %s...\n", hash[:8]+"...")
		infoColor.Printf("   Keeping: %s (%.2f MB, modified: %s)\n", 
			newestFile.Name, 
			float64(newestFile.Size)/1024/1024, 
			newestFile.LastModified.Format("2006-01-02 15:04:05"))

		// Remove all other duplicates
		for _, file := range files {
			if file.Path == newestFile.Path {
				continue
			}

			if dh.DryRun {
				warningColor.Printf("   🗑️  Would remove: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
			} else {
				fmt.Printf("   🗑️  Removing: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
				err := os.Remove(file.Path)
				if err != nil {
					warningColor.Printf("   ⚠️  Failed to remove %s: %v\n", file.Name, err)
					continue
				}
			}
			
			totalRemoved++
			totalSpaceSaved += file.Size
		}
		fmt.Println()
	}

	if totalRemoved > 0 {
		successColor.Printf("✅ Removed %d duplicate files!\n", totalRemoved)
		successColor.Printf("💾 Space saved: %.2f MB\n", float64(totalSpaceSaved)/1024/1024)
	} else {
		fmt.Println("✅ No files were removed.")
	}

	return nil
}

// RemoveDuplicatesInteractive removes duplicate files with interactive selection
func (dh *DuplicateHandler) RemoveDuplicatesInteractive() error {
	if len(dh.Scanner.Duplicates) == 0 {
		fmt.Println("✅ No duplicates found to remove!")
		return nil
	}

	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)
	errorColor := color.New(color.FgRed, color.Bold)

	fmt.Println("🔄 Interactive duplicate removal...")
	fmt.Println("For each set of duplicates, you'll be asked which file to keep.")
	fmt.Println()

	totalRemoved := 0
	totalSpaceSaved := int64(0)

	for hash, files := range dh.Scanner.Duplicates {
		if len(files) < 2 {
			continue
		}

		infoColor.Printf("📋 Found %d duplicates with hash: %s\n", len(files), hash[:8]+"...")
		
		// Display files with numbers
		for i, file := range files {
			fmt.Printf("   %d. %s (%.2f MB, modified: %s)\n", 
				i+1, 
				file.Name, 
				float64(file.Size)/1024/1024, 
				file.LastModified.Format("2006-01-02 15:04:05"))
		}

		// Ask user which file to keep
		var choice int
		for {
			fmt.Printf("\n🤔 Which file would you like to keep? (1-%d, or 0 to skip): ", len(files))
			_, err := fmt.Scanln(&choice)
			if err != nil {
				fmt.Println("   Please enter a valid number.")
				// Clear the input buffer to prevent infinite loop
				var discard string
				fmt.Scanln(&discard)
				continue
			}
			
			if choice == 0 {
				fmt.Println("   Skipping this set of duplicates.")
				break
			}
			
			if choice < 1 || choice > len(files) {
				fmt.Printf("   Please enter a number between 1 and %d.\n", len(files))
				continue
			}
			
			// Valid choice
			keepFile := files[choice-1]
			infoColor.Printf("   Keeping: %s\n", keepFile.Name)
			
			// Remove other files
			for i, file := range files {
				if i == choice-1 {
					continue
				}

				if dh.DryRun {
					warningColor.Printf("   🗑️  Would remove: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
				} else {
					fmt.Printf("   🗑️  Removing: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
					err := os.Remove(file.Path)
					if err != nil {
						errorColor.Printf("   ❌ Failed to remove %s: %v\n", file.Name, err)
						continue
					}
				}
				
				totalRemoved++
				totalSpaceSaved += file.Size
			}
			
			break
		}
		
		fmt.Println()
	}

	if totalRemoved > 0 {
		successColor.Printf("✅ Removed %d duplicate files!\n", totalRemoved)
		successColor.Printf("💾 Space saved: %.2f MB\n", float64(totalSpaceSaved)/1024/1024)
	} else {
		fmt.Println("✅ No files were removed.")
	}

	return nil
}

// RemoveDuplicatesByPattern removes duplicates based on naming patterns
func (dh *DuplicateHandler) RemoveDuplicatesByPattern() error {
	if len(dh.Scanner.Duplicates) == 0 {
		fmt.Println("✅ No duplicates found to remove!")
		return nil
	}

	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	fmt.Println("🔄 Removing duplicates by pattern...")
	fmt.Println("Keeping files without copy indicators like '(1)', 'copy', etc.")
	fmt.Println()

	totalRemoved := 0
	totalSpaceSaved := int64(0)

	for hash, files := range dh.Scanner.Duplicates {
		if len(files) < 2 {
			continue
		}

		// Find the file that looks like the original (no copy indicators)
		var originalFile *FileInfo
		var copyFiles []FileInfo

		for i := range files {
			if dh.isOriginalFile(files[i].Name) {
				originalFile = &files[i]
			} else {
				copyFiles = append(copyFiles, files[i])
			}
		}

		// If we couldn't determine an original, keep the newest
		if originalFile == nil {
			originalFile = &files[0]
			for _, file := range files {
				if file.LastModified.After(originalFile.LastModified) {
					originalFile = &file
				}
			}
			// Add all other files to copies
			for _, file := range files {
				if file.Path != originalFile.Path {
					copyFiles = append(copyFiles, file)
				}
			}
		}

		infoColor.Printf("📋 Processing duplicates for hash: %s...\n", hash[:8]+"...")
		infoColor.Printf("   Keeping: %s (%.2f MB)\n", originalFile.Name, float64(originalFile.Size)/1024/1024)

		// Remove copy files
		for _, file := range copyFiles {
			if dh.DryRun {
				warningColor.Printf("   🗑️  Would remove: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
			} else {
				fmt.Printf("   🗑️  Removing: %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
				err := os.Remove(file.Path)
				if err != nil {
					warningColor.Printf("   ⚠️  Failed to remove %s: %v\n", file.Name, err)
					continue
				}
			}
			
			totalRemoved++
			totalSpaceSaved += file.Size
		}
		fmt.Println()
	}

	if totalRemoved > 0 {
		successColor.Printf("✅ Removed %d duplicate files!\n", totalRemoved)
		successColor.Printf("💾 Space saved: %.2f MB\n", float64(totalSpaceSaved)/1024/1024)
	} else {
		fmt.Println("✅ No files were removed.")
	}

	return nil
}

// isOriginalFile determines if a filename looks like an original (not a copy)
func (dh *DuplicateHandler) isOriginalFile(filename string) bool {
	lowerName := strings.ToLower(filename)
	
	// Patterns that indicate a file is a copy
	copyPatterns := []string{
		" (1)", " (2)", " (3)", " (4)", " (5)", " (6)", " (7)", " (8)", " (9)", " (10)",
		" copy", " copy (1)", " copy (2)", " copy (3)", " copy (4)", " copy (5)",
		" - copy", " - copy (1)", " - copy (2)", " - copy (3)", " - copy (4)", " - copy (5)",
		"_copy", "_copy(1)", "_copy(2)", "_copy(3)", "_copy(4)", "_copy(5)",
		"-copy", "-copy(1)", "-copy(2)", "-copy(3)", "-copy(4)", "-copy(5)",
		".copy", ".copy(1)", ".copy(2)", ".copy(3)", ".copy(4)", ".copy(5)",
		" (copy)", " (copy 1)", " (copy 2)", " (copy 3)", " (copy 4)", " (copy 5)",
		"- (copy)", "- (copy 1)", "- (copy 2)", "- (copy 3)", "- (copy 4)", "- (copy 5)",
		"_ (copy)", "_ (copy 1)", "_ (copy 2)", "_ (copy 3)", "_ (copy 4)", "_ (copy 5)",
		" duplicate", " duplicate (1)", " duplicate (2)", " duplicate (3)", " duplicate (4)", " duplicate (5)",
		" - duplicate", " - duplicate (1)", " - duplicate (2)", " - duplicate (3)", " - duplicate (4)", " - duplicate (5)",
		"_duplicate", "_duplicate(1)", "_duplicate(2)", "_duplicate(3)", "_duplicate(4)", "_duplicate(5)",
		"-duplicate", "-duplicate(1)", "-duplicate(2)", "-duplicate(3)", "-duplicate(4)", "-duplicate(5)",
	}
	
	for _, pattern := range copyPatterns {
		if strings.Contains(lowerName, pattern) {
			return false
		}
	}
	
	return true
}

// MoveDuplicatesToFolder moves duplicate files to a specified folder instead of deleting them
func (dh *DuplicateHandler) MoveDuplicatesToFolder(destFolder string) error {
	if len(dh.Scanner.Duplicates) == 0 {
		fmt.Println("✅ No duplicates found to move!")
		return nil
	}

	successColor := color.New(color.FgGreen, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgCyan)

	// Create destination folder if it doesn't exist
	if !dh.DryRun {
		err := os.MkdirAll(destFolder, 0755)
		if err != nil {
			return fmt.Errorf("failed to create destination folder: %v", err)
		}
	}

	fmt.Printf("🔄 Moving duplicates to: %s\n", destFolder)
	fmt.Println()

	totalMoved := 0
	totalSpaceSaved := int64(0)

	for hash, files := range dh.Scanner.Duplicates {
		if len(files) < 2 {
			continue
		}

		// Find the newest file to keep
		newestFile := files[0]
		for _, file := range files {
			if file.LastModified.After(newestFile.LastModified) {
				newestFile = file
			}
		}

		infoColor.Printf("📋 Processing duplicates for hash: %s...\n", hash[:8]+"...")
		infoColor.Printf("   Keeping: %s (%.2f MB)\n", newestFile.Name, float64(newestFile.Size)/1024/1024)

		// Move all other duplicates
		for _, file := range files {
			if file.Path == newestFile.Path {
				continue
			}

			destPath := filepath.Join(destFolder, file.Name)
			
			if dh.DryRun {
				warningColor.Printf("   📁 Would move: %s -> %s\n", file.Name, destFolder)
			} else {
				fmt.Printf("   📁 Moving: %s\n", file.Name)
				err := dh.atomicMove(file.Path, destPath)
				if err != nil {
					warningColor.Printf("   ⚠️  Failed to move %s: %v\n", file.Name, err)
					continue
				}
			}
			
			totalMoved++
			totalSpaceSaved += file.Size
		}
		fmt.Println()
	}

	if totalMoved > 0 {
		successColor.Printf("✅ Moved %d duplicate files!\n", totalMoved)
		successColor.Printf("💾 Space saved in original folder: %.2f MB\n", float64(totalSpaceSaved)/1024/1024)
	} else {
		fmt.Println("✅ No files were moved.")
	}

	return nil
}

// copyAndDelete copies a file to destination and then deletes the original
func (dh *DuplicateHandler) copyAndDelete(src, dst string) error {
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
	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return err
	}

	// Sync to ensure data is written
	if err := dstFile.Sync(); err != nil {
		return err
	}

	// Delete source file
	return os.Remove(src)
}