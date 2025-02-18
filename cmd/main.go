package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func getDraculaColors() map[string]string {
	draculaColors := map[string]string{
		"Background":  "#282a36",
		"CurrentLine": "#44475a",
		"Selection":   "#44475a",
		"Foreground":  "#f8f8f2",
		"Comment":     "#6272a4",
		"Cyan":        "#8be9fd",
		"Green":       "#50fa7b",
		"Orange":      "#ffb86c",
		"Pink":        "#ff79c6",
		"Purple":      "#bd93f9",
		"Red":         "#ff5555",
		"Yellow":      "#f1fa8c",
	}
	return draculaColors
}

type llm_response struct {
	prompt            string
	response          string
	rendered_response string
	iter              *genai.GenerateContentResponseIterator
	spinner           spinner.Model
	done              bool
}

type StreamingMsg struct {
	err  error
	done bool
}

type StreamingErr struct {
	err error
}

func initialLLMResponse(prompt string) llm_response {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash-lite-preview-02-05")
	if model == nil {
		log.Fatalf("Failed to get generative model")
	}

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	if iter == nil {
		log.Fatalf("Failed to create content stream iterator")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(getDraculaColors()["Selection"]))
	return llm_response{
		prompt:            prompt,
		response:          "",
		rendered_response: "",
		iter:              iter,
		spinner:           s,
		done:              false,
	}
}

func (r llm_response) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return StreamingMsg{err: nil, done: false}
		},
		r.spinner.Tick,
	)
}

func calculate_height(text string) int {
	lines := strings.Split(text, "\n")
	return len(lines) //+ 1
}

func (r llm_response) Streaming() (llm_response, tea.Cmd) {

	resp, err := r.iter.Next()
	if errors.Is(err, iterator.Done) {
		r.done = true
		return r, func() tea.Msg {
			return StreamingMsg{done: true}
		}
	}
	if err != nil {
		log.Error(err)
	}
	for _, cand := range resp.Candidates {
		if len(cand.Content.Parts) > 0 {
			if t, ok := cand.Content.Parts[0].(genai.Text); ok {
				r.response = r.response + string(t)
			}

		}
	}
	out, err := glamour.Render(r.response, "dark")
	_ = err

	r.rendered_response = out

	return r, func() tea.Msg {
		return StreamingMsg{
			err:  nil,
			done: false,
		}
	}

}

func (r llm_response) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(r)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return r, tea.Quit
		}

	case StreamingMsg:
		if msg.err != nil {
			return r, tea.Quit
		}

		if msg.done {
			return r, tea.Quit
		}

		r, streaming_cmd := r.Streaming()

		return r, tea.Batch(
			streaming_cmd,
			func() tea.Msg {
				return tea.WindowSizeMsg{
					// Width:  80,
					Height: calculate_height(r.rendered_response),
				}
			},
		)
	case spinner.TickMsg:
		var spinner_cmd tea.Cmd
		r.spinner, spinner_cmd = r.spinner.Update(msg)
		return r, spinner_cmd
	}

	return r, func() tea.Msg {
		return tea.WindowSizeMsg{
			// Width:  80,
			Height: calculate_height(r.rendered_response),
		}
	}
}

func (r llm_response) View() string {
	if r.done {
		return r.rendered_response
	}
	footer := r.spinner.View() + "Generating..."
	return r.rendered_response + footer
}

func print_banner() {

	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(getDraculaColors()["Green"]))

	// https://patorjk.com/software/taag/#p=display&h=1&v=1&f=Slant&t=PerplxGo
	banner := `
    ____                      __       ______     
   / __ \ ___   _____ ____   / /_  __ / ____/____ 
  / /_/ // _ \ / ___// __ \ / /| |/_// / __ / __ \
 / ____//  __// /   / /_/ // /_>  < / /_/ // /_/ /
/_/     \___//_/   / .___//_//_/|_| \____/ \____/ 
                  /_/                             
`
	fmt.Println(bannerStyle.Render(banner))

	w, _ := lipgloss.Size(banner)

	var style = lipgloss.NewStyle().
		Italic(true).
		Width(w).Align(lipgloss.Center).
		Foreground(lipgloss.Color(getDraculaColors()["Selection"]))

	quote := "Curiosity changes everything."

	fmt.Println(style.Render(quote) + "\n")
}

func ask() {
	var question string

	huh.NewInput().
		Title("ʕ·ᴥ·ʔ Ask anything...").
		Value(&question).
		// WithTheme()
		Run() // this is blocking...

	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		// Margin(0, 2).
		Background(lipgloss.Color(getDraculaColors()["Selection"])).
		Foreground(lipgloss.Color(getDraculaColors()["Foreground"]))

	fmt.Println("⚡" + style.Render(question))

	p := tea.NewProgram(initialLLMResponse(question))

	value, err := p.Run()

	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	_ = value
}

func main() {

	logLevelSting := os.Getenv("LOG_LEVEL")
	switch logLevelSting {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	print_banner()
	ask()
}
