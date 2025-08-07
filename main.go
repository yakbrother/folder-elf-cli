package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// validatePath ensures the path is safe and within allowed directories
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}

	// Check for path traversal attempts
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Get user's home directory for additional validation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %v", err)
	}

	// Ensure path is within user's home directory or system temp
	tempDir := os.TempDir()
	if !strings.HasPrefix(absPath, homeDir) && 
	   !strings.HasPrefix(absPath, tempDir) &&
	   !strings.HasPrefix(absPath, "/tmp") &&
	   !strings.HasPrefix(absPath, "/var/folders") {
		return fmt.Errorf("path must be within user directory or temp directory")
	}

	return nil
}

// getDefaultDownloadsPath returns the default downloads folder path based on the operating system
func getDefaultDownloadsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		// On Windows, try to get the Downloads folder from the registry
		// Fall back to home\Downloads if that fails
		return filepath.Join(home, "Downloads"), nil
	case "darwin":
		// On macOS, the Downloads folder is in the home directory
		return filepath.Join(home, "Downloads"), nil
	case "linux":
		// On Linux, try common locations
		xdgDownloadDir := os.Getenv("XDG_DOWNLOAD_DIR")
		if xdgDownloadDir != "" {
			return xdgDownloadDir, nil
		}
		// Fall back to home/Downloads
		return filepath.Join(home, "Downloads"), nil
	default:
		// For other operating systems, use home/Downloads
		return filepath.Join(home, "Downloads"), nil
	}
}

func main() {
	// Define color schemes for friendly output
	successColor := color.New(color.FgGreen, color.Bold)
	infoColor := color.New(color.FgCyan)
	warningColor := color.New(color.FgYellow)
	errorColor := color.New(color.FgRed, color.Bold)

	app := &cli.App{
		Name:        "elf-cli",
		Usage:       "A friendly tool to clean up your downloads folder",
		Description: "Organize your downloads folder by removing duplicates, categorizing files, and inspecting zip archives.",
		Version:     "1.3.0",
		Authors: []*cli.Author{
			{
				Name: "FolderElf CLI",
			},
		},
		Copyright: "MIT License - see LICENSE file for details",
		UsageText: `elf-cli clean [options]
   elf-cli clean --dry-run --organize --remove-duplicates
   elf-cli clean --path /custom/path --organize-by-date`,
		Commands: []*cli.Command{
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "Clean up your downloads folder",
				Action: func(c *cli.Context) error {
					downloadsPath := c.String("path")
					if downloadsPath == "" {
						// Try to get the default downloads folder
						var err error
						downloadsPath, err = getDefaultDownloadsPath()
						if err != nil {
							errorColor.Printf("❌ Oops! Couldn't find your downloads folder: %v\n", err)
							errorColor.Printf("💡 Please specify a path using --path or -p\n")
							return err
						}
					}

					// Validate the path
					if err := validatePath(downloadsPath); err != nil {
						errorColor.Printf("❌ Invalid path: %v\n", err)
						return err
					}

					infoColor.Printf("🧹 Starting to clean up your downloads folder...\n")
					infoColor.Printf("📂 Looking at: %s\n", downloadsPath)

					// Check if downloads folder exists
					if _, err := os.Stat(downloadsPath); os.IsNotExist(err) {
						errorColor.Printf("❌ Oh no! The downloads folder doesn't exist: %s\n", downloadsPath)
						return fmt.Errorf("downloads folder not found")
					}

					dryRun := c.Bool("dry-run")
					
					// Show prominent warning about destructive operations
					errorColor.Printf("⚠️  WARNING: This tool performs DESTRUCTIVE file operations!\n")
					errorColor.Printf("⚠️  Files may be DELETED or MOVED permanently.\n")
					
					if !dryRun {
						errorColor.Printf("⚠️  Use --dry-run first to preview changes safely.\n")
						fmt.Println()
						
						// Skip confirmation if --force flag is used
						if !c.Bool("force") {
							// Ask for confirmation before proceeding
							fmt.Print("🤔 Do you want to continue? (y/N): ")
							var response string
							fmt.Scanln(&response)
							
							response = strings.ToLower(strings.TrimSpace(response))
							if response != "y" && response != "yes" {
								fmt.Println("❌ Operation cancelled by user.")
								return nil
							}
							fmt.Println()
						} else {
							warningColor.Printf("⚠️  Force mode enabled - skipping confirmation prompt\n")
						}
					}
					
					if dryRun {
						warningColor.Printf("⚠️  Dry run mode enabled - no files will be moved or deleted\n")
					}

					// Create a new scanner and scan the directory
					scanner := NewScanner()
					scanErr := scanner.ScanDirectory(downloadsPath)
					if scanErr != nil {
						errorColor.Printf("❌ Error scanning directory: %v\n", scanErr)
						return scanErr
					}

					// Print the scan results
					scanner.PrintSummary()

					// Handle duplicates if requested
					if c.Bool("remove-duplicates") || c.Bool("interactive-duplicates") || c.Bool("pattern-duplicates") || c.String("move-duplicates") != "" {
						duplicateHandler := NewDuplicateHandler(scanner, dryRun)
						
						if c.Bool("interactive-duplicates") {
							fmt.Println("\n🔄 Starting interactive duplicate removal...")
							err := duplicateHandler.RemoveDuplicatesInteractive()
							if err != nil {
								errorColor.Printf("❌ Error during interactive duplicate removal: %v\n", err)
								return err
							}
						} else if c.Bool("pattern-duplicates") {
							fmt.Println("\n🔄 Starting pattern-based duplicate removal...")
							err := duplicateHandler.RemoveDuplicatesByPattern()
							if err != nil {
								errorColor.Printf("❌ Error during pattern-based duplicate removal: %v\n", err)
								return err
							}
						} else if moveFolder := c.String("move-duplicates"); moveFolder != "" {
							// Validate move folder path
							if err := validatePath(moveFolder); err != nil {
								errorColor.Printf("❌ Invalid move folder path: %v\n", err)
								return err
							}
							fmt.Printf("\n🔄 Moving duplicates to: %s\n", moveFolder)
							err := duplicateHandler.MoveDuplicatesToFolder(moveFolder)
							if err != nil {
								errorColor.Printf("❌ Error moving duplicates: %v\n", err)
								return err
							}
						} else {
							fmt.Println("\n🔄 Starting automatic duplicate removal...")
							err := duplicateHandler.RemoveDuplicates()
							if err != nil {
								errorColor.Printf("❌ Error removing duplicates: %v\n", err)
								return err
							}
						}
					}

					// Handle file organization if requested
					if c.Bool("organize") || c.Bool("organize-by-date") || c.Bool("organize-by-size") || c.Bool("process-zips") {
						organizer := NewFileOrganizer(scanner, dryRun, downloadsPath)
						
						if c.Bool("organize-by-date") {
							fmt.Println("\n📅 Starting date-based organization...")
							err := organizer.OrganizeByDate()
							if err != nil {
								errorColor.Printf("❌ Error during date-based organization: %v\n", err)
								return err
							}
						} else if c.Bool("organize-by-size") {
							fmt.Println("\n📏 Starting size-based organization...")
							err := organizer.OrganizeBySize()
							if err != nil {
								errorColor.Printf("❌ Error during size-based organization: %v\n", err)
								return err
							}
						} else if c.Bool("process-zips") {
							fmt.Println("\n📦 Starting zip file processing...")
							err := organizer.ProcessZipFiles()
							if err != nil {
								errorColor.Printf("❌ Error during zip file processing: %v\n", err)
								return err
							}
						} else {
							fmt.Println("\n📁 Starting file organization by category...")
							err := organizer.OrganizeFiles()
							if err != nil {
								errorColor.Printf("❌ Error during file organization: %v\n", err)
								return err
							}
						}
					}

					successColor.Printf("✨ All done! Your downloads folder is now organized.\n")
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "Path to the downloads folder",
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"d"},
						Usage:   "Show what would be done without actually doing it",
					},
					&cli.BoolFlag{
						Name:    "remove-duplicates",
						Aliases: []string{"r"},
						Usage:   "Remove duplicate files automatically (keeps newest)",
					},
					&cli.BoolFlag{
						Name:    "interactive-duplicates",
						Aliases: []string{"i"},
						Usage:   "Interactively select which duplicate files to keep",
					},
					&cli.BoolFlag{
						Name:    "pattern-duplicates",
						Aliases: []string{"a"},
						Usage:   "Remove duplicates based on naming patterns (keeps files without copy indicators)",
					},
					&cli.StringFlag{
						Name:    "move-duplicates",
						Aliases: []string{"m"},
						Usage:   "Move duplicate files to specified folder instead of deleting",
					},
					&cli.BoolFlag{
						Name:    "organize",
						Aliases: []string{"o"},
						Usage:   "Organize files into category folders (Images, Documents, etc.)",
					},
					&cli.BoolFlag{
						Name:    "organize-by-date",
						Aliases: []string{"od"},
						Usage:   "Organize files into date-based folders (YYYY-MM format)",
					},
					&cli.BoolFlag{
						Name:    "organize-by-size",
						Aliases: []string{"os"},
						Usage:   "Organize files into size-based folders (Tiny, Small, Medium, Large, Huge)",
					},
					&cli.BoolFlag{
						Name:    "process-zips",
						Aliases: []string{"z"},
						Usage:   "Analyze zip file contents and move them to appropriate category folders",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Skip confirmation prompt (useful for automated scripts)",
					},
				},
			},
			{
				Name:    "about",
				Aliases: []string{"a"},
				Usage:   "About this tool",
				Action: func(c *cli.Context) error {
					successColor.Printf("🧝‍♀️ FolderElf CLI - Your friendly downloads folder organizer!\n")
					infoColor.Printf("This tool helps you keep your downloads folder tidy by:\n")
					fmt.Println("  • Removing duplicate files")
					fmt.Println("  • Sorting files into appropriate folders")
					fmt.Println("  • Inspecting zip files to organize their contents")
					return nil
				},
			},
		},
	}

	// Custom help template with friendly colors
	cli.AppHelpTemplate = `{{.Name}} - {{.Usage}}

{{.Version}}

{{if .Commands}}
  📋 Commands:
  {{range .Commands}}
    {{join .Names ", "}}{{"\t"}}{{.Usage}}
  {{end}}{{end}}

{{if .Flags}}
  🚩 Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
`

	if err := app.Run(os.Args); err != nil {
		errorColor.Printf("❌ Something went wrong: %v\n", err)
		log.Fatal(err)
	}
}