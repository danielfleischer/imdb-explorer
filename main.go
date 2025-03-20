package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	// "os"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var apiKey = os.Getenv("OMDB_API_KEY")

type Movie struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	IMDBID string `json:"imdbID"`
	Type   string `json:"Type"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "imdb [movie title]",
		Short: "CLI app to search movies on OMDB",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			title := args[0]
			if apiKey == "" {
				fmt.Println("Error: OMDB_API_KEY environment variable is not set.")
				return
			}

			movies, err := searchOMDB(title)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			displayMovies(movies)
		},
	}

	rootCmd.Execute()
}

func displayMovies(movies []Movie) {
	maxTitleLength := 0
	for _, movie := range movies {
		if len(movie.Title) > maxTitleLength {
			maxTitleLength = len(movie.Title)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 1, 2,' ', 0)
	titleColor := color.New(color.FgCyan).SprintFunc()
	yearColor := color.New(color.FgGreen).SprintFunc()
	linkColor := color.New(color.FgYellow).SprintFunc()

	// fmt.Fprintln(w, "Title\tYear\tType\tIMDB_Link")
	for _, movie := range movies {
		fmt.Fprintf(w,
			"%s\t%s\t%s\t%s\n",
			titleColor(movie.Title),
			yearColor(movie.Year),
			movie.Type,
			linkColor(fmt.Sprintf("https://www.imdb.com/title/%s", movie.IMDBID)))
	}
	w.Flush()
}

func openBrowser(imdbID string) {
	url := fmt.Sprintf("https://www.imdb.com/title/%s", imdbID)
	err := exec.Command("open", url).Start()
	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func searchOMDB(title string) ([]Movie, error) {
	url := fmt.Sprintf("http://www.omdbapi.com/?s=%s&apikey=%s", title, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Search []Movie `json:"Search"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Search, nil
}
