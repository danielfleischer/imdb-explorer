# IMDB Explorer
![GitHub Tag](https://img.shields.io/github/v/tag/danielfleischer/imdb-explorer?color=orange)
[![License](https://img.shields.io/badge/License-Apache%202.0-red.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/danielfleischer/imdb-explorer.svg)](https://pkg.go.dev/github.com/danielfleischer/imdb-explorer)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielfleischer/imdb-explorer)](https://goreportcard.com/report/github.com/danielfleischer/imdb-explorer)
[![Go Version](https://img.shields.io/github/go-mod/go-version/danielfleischer/imdb-explorer?color=00ADD8)](https://golang.org/)

A simple CLI application to search movies and shows using the OMDB API.

<img src="./screenshot.png" alt="image showing how to search shows and movies called 'the matrix'" width="600"/>

## Usage

Run the compiled binary with the movie or show title as an argument, and optionally a year argument `-y`. For example:

```bash
./imdb "The Matrix"
```

> [!TIP]
> Browse with `RET`, toggle info-box with `TAB`, see reviews with `r`, go back with `b`.

-----------

## Prerequisites

- Go installed (tested with go 1.24.1).
- Environment variable `OMDB_API_KEY` set with your [OMDB API](https://www.omdbapi.com/) (free) key.

## Installation

Clone the repository and run:

```bash
go build ./cmd/imdb
```

Alternatively, you can use the provided Makefile targets to manage the project:

- Build the application: `make build`
- Install the application: `make install`

------------

> [!NOTE]  
>
> #### Developed using [Aider](https://aider.chat/)

## TODO

- [x] Jump to reviews (shortcut or a menu).
- [x] Show more info for a movie: rating, genre, director, awards, plot, etc.
- [x] Add general support for opening links; currently using `open`.
- [x] Start a new query.
- [x] Spinner.
- [ ] Show poster (maybe terminal dependent).

## License

For personal use, this project is licensed under the Apache-2.0 License.
