# CLI IMDB Search

A simple CLI application to search movies using the OMDB API.

## Prerequisites

- Go installed (tested with go 1.24.1)
- Environment variable `OMDB_API_KEY` set with your OMDB API key

## Project Structure

```
/ (project root)
├── go.mod
├── go.sum
├── README.md           <-- This file
└── cmd
    └── imdb
         └── main.go    <-- Main CLI application code
```

## Installation

Clone the repository and run:

```bash
go build ./cmd/imdb
```

## Usage

Run the compiled binary with the movie title as an argument. For example:

```bash
./imdb "The Matrix"
```

## License

MIT License
