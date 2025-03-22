# IMDB Explorer

A simple CLI application to search movies and shows using the OMDB API.

<img src="./screenshot.png" alt="image showing how to search shows and movies called 'the matrix'" width="600"/>

## Prerequisites

- Go installed (tested with go 1.24.1).
- Environment variable `OMDB_API_KEY` set with your [OMDB API](https://www.omdbapi.com/) (free) key.

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

Alternatively, you can use the provided Makefile targets to manage the project:
- Build the application: `make build`
- Install the application: `make install`

## License

Apache-2.0 License.
