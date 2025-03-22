# IMDB Explorer

A simple CLI application to search movies and shows using the OMDB API.

<img src="./screenshot.png" alt="image showing how to search shows and movies called 'the matrix'" width="600"/>


## Usage

Run the compiled binary with the movie or show title as an argument, and optionally a year argument `-y`. For example:

```bash
./imdb "The Matrix"
```

Show more information with `TAB`.

> [!NOTE]  
> #### Created using [Aider](https://aider.chat/).

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

## TODO

- [x] Jump to reviews (shortcut or a menu).
- [x] Show more info for a movie: rating, genre, director, awards, plot, etc.
- [ ] Show poster (maybe terminal dependent).

## License

Apache-2.0 License.
