package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

var apiKey = os.Getenv("OMDB_API_KEY")

var (
	titleColor  = color.New(color.FgCyan).SprintFunc()
	yearColor   = color.New(color.FgGreen).SprintFunc()
	ratingColor = color.New(color.FgHiBlue).SprintFunc()
	hintColor   = color.New(color.FgHiBlack).SprintFunc()
)

type Program struct {
	Title   string `json:"Title"`
	Year    string `json:"Year"`
	IMDBID  string `json:"imdbID"`
	Type    string `json:"Type"`
	Seasons string `json:"totalSeasons"`
	Length  string `json:"Runtime"`
	Rating  string `json:"imdbRating"`
}

type Episode struct {
	Title    string `json:"Title"`
	Released string `json:"Released"`
	Episode  string `json:"Episode"`
	ImdbID   string `json:"imdbID"`
	Rating   string `json:"imdbRating"`
}

type EpisodeResponse struct {
	Title    string    `json:"Title"`
	Season   string    `json:"Season"`
	Episodes []Episode `json:"Episodes"`
	Response string    `json:"Response"`
}

type EpisodeRow struct {
	Episode  string
	Title    string
	Rating   string
	Released string
	Link     string
}

type model struct {
	table           table.Model
	state           string
	selectedProgram Program
	movies          []Program
	episodeRows     []EpisodeRow
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
		case "up", "ctrl+p", "p":
			m.table.MoveUp(1)
		case "down", "ctrl+n", "n":
			m.table.MoveDown(1)
		case "r":
			var link string
			switch m.state {
			case "", "default":
				movie := m.movies[m.table.Cursor()]
				link = fmt.Sprintf("https://www.imdb.com/title/%s", movie.IMDBID)
			case "episodeDisplay":
				link = m.episodeRows[m.table.Cursor()].Link
			default:
				movie := m.movies[m.table.Cursor()]
				link = fmt.Sprintf("https://www.imdb.com/title/%s", movie.IMDBID)
			}
			openBrowser(fmt.Sprintf("%s/reviews", link))
			return m, tea.Quit
		case "enter":
			if m.state == "" || m.state == "default" {
				movie := m.movies[m.table.Cursor()]
				if movie.Type == "series" {
					m.selectedProgram = movie
					columns := []table.Column{
						{Title: "Options", Width: 20},
					}
					optionsRows := []table.Row{
						{"Find Episode"},
						{"Browse"},
					}
					m.table = table.New(
						table.WithColumns(columns),
						table.WithRows(optionsRows),
						table.WithFocused(true),
					)
					m.table.SetHeight(len(optionsRows) + 2)
					m.state = "episodeOptions"
					return m, nil
				} else {
					openBrowser(fmt.Sprintf("https://www.imdb.com/title/%s", movie.IMDBID))
					return m, tea.Quit
				}
			} else if m.state == "episodeOptions" {
				selectedOption := m.table.SelectedRow()[0]
				if selectedOption == "Browse" {
					openBrowser(fmt.Sprintf("https://www.imdb.com/title/%s", m.selectedProgram.IMDBID))
					return m, tea.Quit
				} else if selectedOption == "Find Episode" {
					count, err := strconv.Atoi(m.selectedProgram.Seasons)
					if err != nil {
						fmt.Println("Invalid seasons count:", m.selectedProgram.Seasons)
						return m, tea.Quit
					}
					columns := []table.Column{
						{Title: "Season", Width: 10},
					}
					var seasonRows []table.Row
					for i := 1; i <= count; i++ {
						seasonRows = append(seasonRows, table.Row{fmt.Sprintf("%d", i)})
					}
					m.table = table.New(
						table.WithColumns(columns),
						table.WithRows(seasonRows),
						table.WithFocused(true),
					)
					m.table.SetHeight(len(seasonRows) + 2)
					m.state = "seasonSelection"
					return m, nil
				}
			} else if m.state == "seasonSelection" {
				selectedSeason := m.table.SelectedRow()[0]
				episodes, err := getEpisodes(m.selectedProgram.IMDBID, selectedSeason)
				if err != nil {
					fmt.Println("Error fetching episodes:", err)
					return m, tea.Quit
				}
				maxTitleLength := 0
				for _, episode := range episodes {
					if len(episode.Title) > maxTitleLength {
						maxTitleLength = len(episode.Title)
					}
				}
				columns := []table.Column{
					{Title: "Episode", Width: 10},
					{Title: "Title", Width: maxTitleLength + 10},
					{Title: "Rating", Width: 12},
					{Title: "Released", Width: 20},
					{Title: "Link", Width: 0},
				}
				var tableRows []table.Row
				var episodeRows []EpisodeRow
				for _, ep := range episodes {
					episodeRows = append(episodeRows, EpisodeRow{
						Episode:  ep.Episode,
						Title:    ep.Title,
						Rating:   ep.Rating,
						Released: ep.Released,
						Link:     fmt.Sprintf("https://www.imdb.com/title/%s", ep.ImdbID),
					})
					tableRows = append(tableRows, table.Row{
						ep.Episode,
						titleColor(ep.Title),
						ratingColor(ep.Rating),
						yearColor(ep.Released),
						fmt.Sprintf("https://www.imdb.com/title/%s", ep.ImdbID),
					})
				}
				m.table = table.New(
					table.WithColumns(columns),
					table.WithRows(tableRows),
					table.WithFocused(true),
				)
				m.table.SetHeight(len(tableRows) + 2)
				m.episodeRows = episodeRows
				m.state = "episodeDisplay"
				return m, nil
			} else if m.state == "episodeDisplay" {
				link := m.episodeRows[m.table.Cursor()].Link
				openBrowser(link)
				return m, tea.Quit
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	return m.table.View() + hintColor("\n\nRET: browse, r: reviews, up/down/n/p: move, q: quit")
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

			for i, movie := range movies {
				info, err := getProgramInfo(movie.IMDBID)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				movies[i].Seasons = info.Seasons
				movies[i].Length = info.Length
				movies[i].Rating = info.Rating
			}

			displayMovies(movies)
		},
	}

	rootCmd.Flags().StringVarP(&year, "year", "y", "", "Year of release")

	rootCmd.Execute()
}

func displayMovies(movies []Program) {
	maxTitleLength := 0
	for _, movie := range movies {
		if len(movie.Title) > maxTitleLength {
			maxTitleLength = len(movie.Title)
		}
	}

	columns := []table.Column{
		{Title: "Title", Width: maxTitleLength + 10},
		{Title: "Year", Width: 16},
		{Title: "Rating", Width: 12},
		{Title: "Type", Width: 10},
		{Title: "Length", Width: 10},
		{Title: "Seasons", Width: 10},
		{Title: "Link", Width: 0},
	}

	var rows []table.Row
	for _, item := range movies {
		rows = append(rows, table.Row{
			titleColor(item.Title),
			yearColor(item.Year),
			ratingColor(item.Rating),
			cases.Title(language.English).String(item.Type),
			item.Length,
			item.Seasons,
			fmt.Sprintf("https://www.imdb.com/title/%s", item.IMDBID),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	t.SetHeight(len(rows) + 2)

	m := model{table: t, movies: movies}

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

func searchOMDB(title, year string) ([]Program, error) {
	url := fmt.Sprintf("https://www.omdbapi.com/?s=%s&y=%s&apikey=%s", title, year, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Search []Program `json:"Search"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Search, nil
}

func getProgramInfo(imdbID string) (Program, error) {
	url := fmt.Sprintf("https://www.omdbapi.com/?i=%s&apikey=%s", imdbID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return Program{}, err
	}
	defer resp.Body.Close()

	var info Program
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return Program{}, err
	}

	return info, nil
}

func getEpisodes(imdbID, season string) ([]Episode, error) {
	url := fmt.Sprintf("https://www.omdbapi.com/?i=%s&Season=%s&apikey=%s", imdbID, season, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var epResponse EpisodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&epResponse); err != nil {
		return nil, err
	}
	return epResponse.Episodes, nil
}
