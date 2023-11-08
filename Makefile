# Variables
BINARY_NAME=mm-weather
GO=go

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME)_linux_amd64

# Build for Linux
build-mac:
	@echo "Building for Linux..."
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BINARY_NAME)_macos_arm64

# Clean Up
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)_*
