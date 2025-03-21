package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/exec"
)

var apiKey = os.Getenv("OMDB_API_KEY")

type Movie struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	IMDBID string `json:"imdbID"`
	Type   string `json:"Type"`
}

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.table.MoveUp(1)
		case "down":
			m.table.MoveDown(1)
		case "enter":
			selectedRow := m.table.SelectedRow()
			link := string(selectedRow[3])
			openBrowser(link)
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return m.table.View()
}

func main() {
	var year string

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

			movies, err := searchOMDB(title, year)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			displayMovies(movies)
		},
	}

	rootCmd.Flags().StringVarP(&year, "year", "y", "", "Year of release")

	rootCmd.Execute()
}

func displayMovies(movies []Movie) {
	maxTitleLength := 0
	for _, movie := range movies {
		if len(movie.Title) > maxTitleLength {
			maxTitleLength = len(movie.Title)
		}
	}

	columns := []table.Column{
		{Title: "Title", Width: maxTitleLength + 10},
		{Title: "Year", Width: 18},
		{Title: "Type", Width: 10},
		{Title: "Link", Width: 0},
	}

	titleColor := color.New(color.FgCyan).SprintFunc()
	yearColor := color.New(color.FgGreen).SprintFunc()
	// linkColor := color.New(color.FgYellow).SprintFunc()

	var rows []table.Row
	for _, movie := range movies {
		rows = append(rows, table.Row{
			titleColor(movie.Title),
			yearColor(movie.Year),
			movie.Type,
			fmt.Sprintf("https://www.imdb.com/title/%s", movie.IMDBID),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	m := model{table: t}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}

}

func openBrowser(url string) {
	err := exec.Command("open", url).Start()
	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func searchOMDB(title, year string) ([]Movie, error) {
	url := fmt.Sprintf("http://www.omdbapi.com/?s=%s&y=%s&apikey=%s", title, year, apiKey)
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
