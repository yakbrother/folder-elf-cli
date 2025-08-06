# FolderElf CLI - A friendly tool to clean up your downloads folder

FolderElf CLI (command name: elf-cli) is a command-line tool written in Go that helps you organize your downloads folder by removing duplicates, categorizing files, and inspecting zip archives to determine their appropriate folders.

## ⚠️ Security Notice

**This tool performs destructive file operations!** It can delete and move files permanently. Always use `--dry-run` first to preview what will be done before running actual operations.

## Features

- **Duplicate Detection and Removal**: Finds duplicate files using MD5 hashes and removes them, keeping only the newest version
- **File Organization**: Automatically sorts files into categorized folders (Images, Documents, Videos, etc.)
- **Zip File Inspection**: Examines the contents of zip files to determine their appropriate category
- **Multiple Organization Strategies**: Organize by category, date (YYYY-MM format), or file size
- **Dry Run Mode**: Preview what would be done without actually making any changes
- **Friendly Colored Output**: Easy-to-read output with colors and emojis
- **Security Features**: Path validation, zip bomb protection, and atomic file operations

## Prerequisites

- **Go 1.20 or later** (required for building)
- Git (for cloning the repository)

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Pre-built binaries are available for:

- **macOS**: Intel (AMD64) and Apple Silicon (ARM64)
- **Linux**: AMD64, ARM64, and ARM32
- **Windows**: AMD64 and ARM64

Download from the [Releases](https://github.com/yakbrother/folder-elf-cli/releases) section.

**Quick install (Linux/macOS):**

```bash
# Download and install in one command
curl -L https://github.com/yakbrother/folder-elf-cli/releases/latest/download/elf-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o elf-cli && chmod +x elf-cli
```

### Option 2: Build from Source

1. **Clone this repository**:

   ```bash
   git clone https://github.com/yakbrother/folder-elf-cli.git
   cd folder-elf-cli
   ```

2. **Build the tool**:

   ```bash
   go build -o elf-cli
   ```

3. **Make it executable** (optional):

   ```bash
   chmod +x elf-cli
   ```

4. **Move it to a directory in your PATH** (optional):
   ```bash
   mv elf-cli /usr/local/bin/
   ```

### Option 3: Install via Go

If you have Go installed:

```bash
go install github.com/yakbrother/folder-elf-cli@latest
```

This will install the tool to `$GOPATH/bin/` or `$GOBIN/`.

## Usage

### Basic Usage

To scan your downloads folder and see what's there:

```bash
./elf-cli clean
```

This will scan your downloads folder and show you what it found, but won't make any changes unless you add specific flags.

**To actually organize files, use:**

```bash
./elf-cli clean --organize --remove-duplicates
```

### How the Tool Finds Your Downloads Folder

The tool automatically detects your downloads folder based on your operating system:

- **macOS**: `/Users/[username]/Downloads`
- **Windows**: `C:\Users\[username]\Downloads`
- **Linux**: Uses the `XDG_DOWNLOAD_DIR` environment variable if set, otherwise `/home/[username]/Downloads`

### Using the Tool with Any Folder

You can use the tool with any folder by specifying the path:

```bash
./elf-cli clean --path /path/to/any/folder
```

This makes elf-cli versatile for organizing any directory, not just downloads folders.

### Dry Run Mode

To see what would be done without actually making any changes:

```bash
./elf-cli clean --dry-run
```

**Note**: The tool will show a warning and ask for confirmation before making changes. Use `--force` to skip the confirmation prompt (useful for automated scripts).

### Removing Duplicates

To automatically remove duplicate files (keeping the newest version):

```bash
./elf-cli clean --remove-duplicates
```

Other duplicate removal options:

- `--interactive-duplicates`: Interactively select which duplicate files to keep
- `--pattern-duplicates`: Remove duplicates based on naming patterns (keeps files without copy indicators like "(1)", "(2)", "copy", etc.)
- `--move-duplicates <folder>`: Move duplicate files to a specified folder instead of deleting them

### Organizing Files

To organize files into category folders:

```bash
./elf-cli clean --organize
```

Other organization options:

- `--organize-by-date`: Organize files into date-based folders (YYYY-MM format)
- `--organize-by-size`: Organize files into size-based folders (Tiny, Small, Medium, Large, Huge)

### Processing Zip Files

To analyze zip file contents and move them to appropriate category folders:

```bash
./elf-cli clean --process-zips
```

### Combining Options

You can combine multiple options:

```bash
./elf-cli clean --dry-run --organize --remove-duplicates --process-zips
```

**Available flags:**

- `--dry-run` - Preview changes without making them
- `--force` - Skip confirmation prompt (for automation)
- `--organize` - Organize files by category
- `--organize-by-date` - Organize files by date
- `--organize-by-size` - Organize files by size
- `--remove-duplicates` - Remove duplicate files
- `--pattern-duplicates` - Remove duplicates by naming patterns
- `--interactive-duplicates` - Interactive duplicate removal
- `--move-duplicates <folder>` - Move duplicates to folder
- `--process-zips` - Process zip file contents
- `--path <path>` - Specify custom folder path

### Specifying a Custom Path

By default, elf-cli looks at your downloads folder. To specify a different path:

```bash
./elf-cli clean --path /path/to/your/folder
```

## File Categories

Files are organized into the following categories:

- **Images**: JPG, PNG, GIF, SVG, WebP, HEIC, and other image formats
- **Documents**: PDF, DOC, DOCX, TXT, MD, EPUB, and other document formats
- **Videos**: MP4, MOV, AVI, MKV, and other video formats
- **Music**: MP3, WAV, FLAC, AAC, and other audio formats
- **Archives**: ZIP, RAR, 7Z, TAR, GZ, and other archive formats
- **Disk Images**: DMG, ISO, and other disk image formats
- **Applications**: APP, EXE, and other application formats
- **Other**: Files that don't fit into any of the above categories

## How Organization Works

### Organization by Category

Files are moved into folders based on their type (Images, Documents, Videos, etc.). The files themselves are not renamed, just moved to the appropriate category folder. For example:

- `image.jpg` → `Images/image.jpg`
- `document.pdf` → `Documents/document.pdf`

### Organization by Date

Files are moved into folders based on their last modification date (in YYYY-MM format). The files themselves are not renamed, just organized into these date-based folders. For example:

- A file modified on August 6, 2025 → `2025-08/filename.ext`
- A file modified on December 25, 2024 → `2024-12/filename.ext`

### Organization by Size

Files are moved into folders based on their file size. The files themselves are not renamed, just moved to the appropriate size-based folder. For example:

- A 500KB file → `Tiny/filename.ext`
- A 5MB file → `Small/filename.ext`
- A 50MB file → `Medium/filename.ext`
- A 500MB file → `Large/filename.ext`
- A 2GB file → `Huge/filename.ext`

## Size Categories

When organizing by size, files are categorized as:

- **Tiny**: Less than 1 MB
- **Small**: 1 MB to 10 MB
- **Medium**: 10 MB to 100 MB
- **Large**: 100 MB to 1 GB
- **Huge**: More than 1 GB

## Examples

1. **Preview what would be done**:

   ```bash
   ./elf-cli clean --dry-run
   ```

2. **Remove duplicates and organize files**:

   ```bash
   ./elf-cli clean --remove-duplicates --organize
   ```

3. **Organize by date and process zip files**:

   ```bash
   ./elf-cli clean --organize-by-date --process-zips
   ```

4. **Move duplicates to a specific folder instead of deleting them**:
   ```bash
   ./elf-cli clean --move-duplicates ~/duplicates-backup
   ```

## Handling Files with "(1)" or "(2)" in the Filename

When you download the same file multiple times, browsers and download managers often add indicators like "(1)", "(2)", "copy", etc. to the filename. elf-cli has a smart duplicate detection feature that can identify these patterns and keep the original file.

### Pattern-Based Duplicate Removal

Using the `--pattern-duplicates` flag, elf-cli will:

1. **Identify copy patterns**: Recognizes files with patterns like:

   - " (1)", " (2)", " (3)", etc.
   - " copy", " copy (1)", " copy (2)", etc.
   - " - copy", " - copy (1)", " - copy (2)", etc.
   - "\_copy", "\_copy(1)", "\_copy(2)", etc.
   - "-copy", "-copy(1)", "-copy(2)", etc.
   - " duplicate", " duplicate (1)", " duplicate (2)", etc.
   - And many other variations

2. **Keep the original**: When duplicates are found, it keeps the file without these copy indicators (the "original") and removes the ones that do.

3. **Fallback to newest**: If it can't determine which is the original (e.g., all files have copy indicators), it falls back to keeping the newest file.

### Example

If you have these files:

- `document.pdf`
- `document (1).pdf`
- `document (2).pdf`

Using `--pattern-duplicates` will keep `document.pdf` and remove the others.

## Running the Tool Repeatedly

elf-cli is designed to be run repeatedly as your downloads folder (or any folder) gets refilled with new files. Each time you run it:

1. **Scans the current state**: It looks at all files currently in the folder
2. **Identifies new duplicates**: Any duplicate files that have been added since the last run will be detected
3. **Organizes new files**: Files that haven't been organized yet will be sorted into the appropriate folders
4. **Processes new archives**: Any new zip files will be inspected and categorized based on their contents

### Example Workflow

1. **Initial cleanup**:

   ```bash
   ./elf-cli clean --dry-run --organize --remove-duplicates --process-zips
   ```

   (Review what will be done)

2. **Actual cleanup**:

   ```bash
   ./elf-cli clean --organize --remove-duplicates --process-zips
   ```

   (Apply the changes)

3. **Weekly maintenance**:
   ```bash
   ./elf-cli clean --organize --remove-duplicates
   ```
   (Run this regularly to keep the folder organized)

## Safety Features

- **Dry Run Mode**: Always preview changes before applying them
- **Path Validation**: Prevents path traversal attacks and restricts operations to safe directories
- **Zip Bomb Protection**: Detects and prevents zip bomb attacks
- **Atomic Operations**: Uses atomic file operations to prevent data corruption
- **Confirmation Prompts**: The tool will ask for confirmation before making destructive changes
- **Detailed Logging**: See exactly what files are being moved or deleted
- **Error Handling**: The tool handles errors gracefully and continues processing other files

## Setting Up as a Cron Job

You can automate the cleanup of your downloads folder by setting up elf-cli as a cron job. This allows the tool to run automatically at scheduled intervals.

### Linux/macOS

1. **Make sure elf-cli is in your PATH**:

   ```bash
   sudo mv elf-cli /usr/local/bin/
   ```

2. **Open your crontab for editing**:

   ```bash
   crontab -e
   ```

3. **Add a cron job entry**. Here are some examples:

   **Run every day at 2 AM**:

   ```
   0 2 * * * /usr/local/bin/elf-cli clean --organize --remove-duplicates --pattern-duplicates --force
   ```

   **Run every Friday at 6 PM**:

   ```
   0 18 * * 5 /usr/local/bin/elf-cli clean --organize --remove-duplicates --pattern-duplicates --force
   ```

   **Run every hour** (for very active download folders):

   ```
   0 * * * * /usr/local/bin/elf-cli clean --organize --remove-duplicates --pattern-duplicates --force
   ```

4. **Save and exit** the editor.

### Windows

1. **Open Task Scheduler**:

   - Press Windows Key and type "Task Scheduler"
   - Click on "Create Basic Task" in the Actions pane

2. **Set up the task**:

   - **Name**: "elf-cli Downloads Cleanup"
   - **Trigger**: Choose how often you want it to run (daily, weekly, etc.)
   - **Action**: "Start a program"
   - **Program/script**: Path to your elf-cli executable
   - **Add arguments**: `clean --organize --remove-duplicates --pattern-duplicates --force`

3. **Finish the wizard** to create the task.

### Logging Cron Job Output

To keep a log of what the cron job does, you can redirect the output to a log file:

```
0 2 * * * /usr/local/bin/elf-cli clean --organize --remove-duplicates --pattern-duplicates --force >> /home/username/elf-cli.log 2>&1
```

### Dry Run in Cron

If you want to test the cron job without actually making changes, use the `--dry-run` flag:

```
0 2 * * * /usr/local/bin/elf-cli clean --dry-run --organize --remove-duplicates --pattern-duplicates --force >> /home/username/elf-cli.log 2>&1
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
