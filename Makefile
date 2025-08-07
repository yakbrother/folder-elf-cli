.PHONY: build test clean install

# Build the application
build:
	go build -o elf-cli

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run specific test files
test-main:
	go test -v main_test.go main.go

test-scanner:
	go test -v scanner_test.go scanner.go

test-duplicates:
	go test -v duplicates_test.go duplicates.go

test-organizer:
	go test -v organizer_test.go organizer.go

# Clean build artifacts
clean:
	rm -f elf-cli
	rm -f coverage.out coverage.html

# Install to system
install: build
	sudo mv elf-cli /usr/local/bin/

# Run with dry-run mode
dry-run:
	./elf-cli clean --dry-run --organize --remove-duplicates

# Create a test directory structure
test-setup:
	mkdir -p test-downloads/{Images,Documents,Videos,Music,Archives,Applications,"Disk Images",Other}
	touch test-downloads/image.jpg
	touch test-downloads/document.pdf
	touch test-downloads/video.mp4
	touch test-downloads/music.mp3
	touch test-downloads/archive.zip
	touch test-downloads/app.exe
	touch test-downloads/disk.iso

# Run integration test
integration-test: test-setup
	./elf-cli clean --path test-downloads --dry-run --organize --remove-duplicates
	rm -rf test-downloads

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o elf-cli-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o elf-cli-linux-arm64
	GOOS=darwin GOARCH=amd64 go build -o elf-cli-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o elf-cli-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o elf-cli-windows-amd64.exe
	GOOS=windows GOARCH=arm64 go build -o elf-cli-windows-arm64.exe
	chmod +x elf-cli-linux-*
	chmod +x elf-cli-darwin-*

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-main      - Run main tests"
	@echo "  test-scanner   - Run scanner tests"
	@echo "  test-duplicates- Run duplicate handler tests"
	@echo "  test-organizer - Run organizer tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to system"
	@echo "  dry-run        - Run with dry-run mode"
	@echo "  test-setup     - Create test directory structure"
	@echo "  integration-test- Run integration test"
	@echo "  build-all      - Build for all platforms"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help" 