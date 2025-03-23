# Default target builds the CLI application.
.PHONY: all build install clean help

VERSION = $(shell git describe --tags --always || echo "dev")

all: build

# Build the application binary and place it in ./bin
build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" ./cmd/imdb

# Install the application binary to your $GOPATH/bin or $HOME/go/bin
install:
	go install -ldflags="-X 'main.Version=$(VERSION)'" ./cmd/imdb

clean:
	rm -f imdb

# Display available make targets
help:
	@echo "Available targets:"
	@echo "  build   - Build the application binary in ./bin"
	@echo "  install - Install the application binary"
	@echo "  clean   - Clean the project"
	@echo "  help    - Display this help message"
