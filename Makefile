# Default target builds the CLI application.
.PHONY: all build install help

all: build

# Build the application binary and place it in ./bin
build:
	mkdir -p bin
	go build -o bin/imdb ./cmd/imdb

# Install the application binary to your $GOPATH/bin or $HOME/go/bin
install:
	go install ./cmd/imdb


# Display available make targets
help:
	@echo "Available targets:"
	@echo "  build   - Build the application binary in ./bin"
	@echo "  install - Install the application binary"
	@echo "  help    - Display this help message"
