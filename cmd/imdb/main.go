package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

var apiKey = os.Getenv("OMDB_API_KEY")
var Version = "dev"

var (
	titleColor  = color.New(color.FgCyan).SprintFunc()
	yearColor   = color.New(color.FgGreen).SprintFunc()
	ratingColor = color.New(color.FgHiBlue).SprintFunc()
	infoColor   = color.New(color.FgYellow).SprintFunc()
	hintColor   = color.New(color.FgHiBlack, color.Bold).SprintFunc()
)

type Program struct {
	Title    string `json:"Title"`
	Year     string `json:"Year"`
	IMDBID   string `json:"imdbID"`
	Type     string `json:"Type"`
	Seasons  string `json:"totalSeasons"`
	Length   string `json:"Runtime"`
	Episode  string `json:"Episode"`
	Rating   string `json:"imdbRating"`
	Released string `json:"Released"`
	Genre    string `json:"Genre"`
	Director string `json:"Director"`
	Actors   string `json:"Actors"`
	Plot     string `json:"Plot"`
	Awards   string `json:"Awards"`
}

func newSearchCmd(query string, year string) tea.Cmd {
	return func() tea.Msg {
		imdbIDs, err := searchOMDB(query, year)
		if err != nil {
			return searchResultsMsg{err: err}
		}
		var programs []Program
		for _, id := range imdbIDs {
			program, err := getProgramInfo(id)
			if err != nil {
				return searchResultsMsg{err: err}
			}
			programs = append(programs, program)
		}
		return searchResultsMsg{programs: programs}
	}
}

type EpisodeResponse struct {
	Title    string    `json:"Title"`
	Season   string    `json:"Season"`
	Episodes []Program `json:"Episodes"`
	Response string    `json:"Response"`
}

type detailsMsg struct {
	details Program
	err     error
	imdbID  string
}
type searchResultsMsg struct {
	programs []Program
	err      error
}

type model struct {
	table           table.Model
	state           string
	selectedProgram Program
	programs        []Program
	episodeRows     []Program
	infoViewport    viewport.Model
	showDetails     bool
	programCache    map[string]Program
	newQueryInput   textinput.Model
	spinner         spinner.Model
	query           string
	year            string
}

func (m model) Init() tea.Cmd {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	m.spinner = s
	return tea.Batch(newSearchCmd(m.query, m.year), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultsMsg:
		if msg.err != nil {
			return m, tea.Quit
		}
		m.programs = msg.programs
		m.table = builtShowsTable(msg.programs)
		m.state = ""
		return m, nil
	case detailsMsg:
		m.programCache[msg.imdbID] = msg.details
		m.infoViewport.SetContent(buildDetails(msg.details))
		return m, nil
	}
	if m.state == "newQuery" {
		var cmd tea.Cmd
		m.newQueryInput, cmd = m.newQueryInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			query := m.newQueryInput.Value()
			m.state = "loading"
			s := spinner.New()
			s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
			m.spinner = s
			return m, tea.Batch(newSearchCmd(query, m.year), m.spinner.Tick)
		}
		return m, cmd
	}
	if m.state == "loading" {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.spinner.Tick)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "ctrl+p", "p":
			m.table.MoveUp(1)
			return maybeUpdateDetails(m)
		case "down", "ctrl+n", "n":
			m.table.MoveDown(1)
			return maybeUpdateDetails(m)
		case "tab":
			m.showDetails = !m.showDetails
			if m.showDetails && (m.state == "" || m.state == "default" || m.state == "episodeDisplay") {
				var imdbID string
				if m.state == "" || m.state == "default" {
					imdbID = m.programs[m.table.Cursor()].IMDBID
				} else {
					imdbID = m.episodeRows[m.table.Cursor()].IMDBID
				}
				vp := viewport.New(100, 10)
				vp.SetContent("")
				m.infoViewport = vp
				if program, ok := m.programCache[imdbID]; ok {
					m.infoViewport.SetContent(buildDetails(program))
					return m, nil
				}
				return m, fetchDetailsCmd(imdbID)
			}
		case "s":
			ti := textinput.New()
			ti.Placeholder = "Enter new query"
			ti.Focus()
			ti.CharLimit = 156
			m.newQueryInput = ti
			m.episodeRows = nil
			m.state = "newQuery"
			return m, nil
		case "r":
			var id string
			switch m.state {
			case "", "default":
				id = m.programs[m.table.Cursor()].IMDBID
			case "episodeDisplay":
				id = m.episodeRows[m.table.Cursor()].IMDBID
			}
			if id != "" {
				openBrowser(fmt.Sprintf("https://www.imdb.com/title/%s/reviews", id))
				return m, tea.Quit
			}
		case "b":
			switch m.state {
			case "episodeDisplay":
				// Go back one level: from episodeDisplay to seasonSelection.
				return builtSeasonSelection(m)
			case "seasonSelection", "episodeOptions":
				// Go back one level: from seasonSelection (or episodeOptions) to main modal.
				{
					m.table = builtShowsTable(m.programs)
					m.state = ""
					if m.showDetails {
						imdbID := m.programs[m.table.Cursor()].IMDBID
						if program, ok := m.programCache[imdbID]; ok {
							m.infoViewport.SetContent(buildDetails(program))
							return m, nil
						}
						return m, fetchDetailsCmd(imdbID)
					}
					return m, nil
				}
			default:
				return m, nil
			}
		case "enter":
			if m.state == "" || m.state == "default" {
				movie := m.programs[m.table.Cursor()]
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
					return builtSeasonSelection(m)
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
				var episodeRows []Program
				for _, ep := range episodes {
					episodeRows = append(episodeRows, ep)
					tableRows = append(tableRows, table.Row{
						ep.Episode,
						titleColor(ep.Title),
						ratingColor(ep.Rating),
						yearColor(ep.Released),
						fmt.Sprintf("https://www.imdb.com/title/%s", ep.IMDBID),
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
				if m.showDetails {
					return m, fetchDetailsCmd(m.episodeRows[m.table.Cursor()].IMDBID)
				}
				return m, nil
			} else if m.state == "episodeDisplay" {
				imdbID := m.episodeRows[m.table.Cursor()].IMDBID
				openBrowser(fmt.Sprintf("https://www.imdb.com/title/%s", imdbID))
				return m, tea.Quit
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.state == "newQuery" {
		return m.newQueryInput.View() + "\n\nPress enter to search."
	}
	if m.state == "loading" {
		return "Searching... " + m.spinner.View()
	}
	view := ""
	view += m.table.View() + hintColor("\n\nRET: browse | r: reviews | TAB: toggle details | up/down/n/p: move | b: back | s: new query | q: quit")
	if m.showDetails {
		view += infoColor("\n================ DETAILS =================\n\n")
		view += infoColor(m.infoViewport.View() + "\n")
		view += infoColor("=========================================\n")
	}
	return view
}

func main() {
	var year string

	var rootCmd = &cobra.Command{
		Use:     "imdb [movie title]",
		Short:   "CLI app to search movies on OMDB",
		Args:    cobra.MinimumNArgs(1),
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			title := args[0]
			if apiKey == "" {
				fmt.Println("Error: OMDB_API_KEY environment variable is not set.")
				return
			}

			s := spinner.New()
			s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))

			m := model{
				table:        builtShowsTable([]Program{}),
				programs:     []Program{},
				programCache: make(map[string]Program),
				query:        title,
				year:         year,
				state:        "loading",
				spinner:      s,
			}

			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&year, "year", "y", "", "Year of release")

	rootCmd.Execute()
}

func builtSeasonSelection(m model) (model, tea.Cmd) {
	count, err := strconv.Atoi(m.selectedProgram.Seasons)
	if err != nil {
		fmt.Println("Invalid seasons count:", m.selectedProgram.Seasons)
		return m, tea.Quit
	}
	columns := []table.Column{
		{Title: "Select Season", Width: 16},
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
	if m.showDetails {
		return m, fetchDetailsCmd(m.selectedProgram.IMDBID)
	}
	return m, nil
}

func builtShowsTable(programs []Program) table.Model {
	maxTitleLength := 0
	for _, movie := range programs {
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
	for _, item := range programs {
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
	return t
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		fmt.Println("Unsupported OS:", runtime.GOOS)
		os.Exit(1)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func searchOMDB(title string, year string) ([]string, error) {
	url := fmt.Sprintf("https://www.omdbapi.com/?s=%s&y=%s&apikey=%s", title, year, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Search []struct {
			ImdbID string `json:"imdbID"`
		} `json:"Search"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var IMDBIDs []string
	for _, item := range result.Search {
		IMDBIDs = append(IMDBIDs, item.ImdbID)
	}

	return IMDBIDs, nil
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

func getEpisodes(imdbID, season string) ([]Program, error) {
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

func maybeUpdateDetails(m model) (model, tea.Cmd) {
	if m.showDetails && (m.state == "" || m.state == "default" || m.state == "episodeDisplay") {
		var imdbID string
		if m.state == "" || m.state == "default" {
			imdbID = m.programs[m.table.Cursor()].IMDBID
		} else {
			imdbID = m.episodeRows[m.table.Cursor()].IMDBID
		}
		if program, ok := m.programCache[imdbID]; ok {
			m.infoViewport.SetContent(buildDetails(program))
			return m, nil
		}
		return m, fetchDetailsCmd(imdbID)
	}
	return m, nil
}

func fetchDetailsCmd(imdbID string) tea.Cmd {
	return func() tea.Msg {
		program, err := getProgramInfo(imdbID)
		return detailsMsg{details: program, err: err, imdbID: imdbID}
	}
}

func buildDetails(program Program) string {
	return fmt.Sprintf(
		"Title: %s\nYear: %s\nRated: %s\nGenre: %s\nDirector: %s\nActors: %s\nPlot: %s\nAwards: %s",
		program.Title,
		program.Year,
		program.Rating,
		program.Genre,
		program.Director,
		program.Actors,
		program.Plot,
		program.Awards,
	)
}
