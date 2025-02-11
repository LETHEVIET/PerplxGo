package main

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	prompt   string
	response string
	iter     *genai.GenerateContentResponseIterator
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

	_ = err

	model := client.GenerativeModel("gemini-pro") // Or your desired model
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	return llm_response{
		prompt:   prompt,
		response: "",
		iter:     iter,
	}
}

func (r llm_response) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return func() tea.Msg {
		return StreamingMsg{err: nil, done: false}
	}
}

func (r llm_response) Streaming() (tea.Model, tea.Cmd) {

	resp, err := r.iter.Next()
	if errors.Is(err, iterator.Done) {
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
				// fmt.Println("current llm resp", string(t))
				// fmt.Println("current r.response", r.response)
				r.response = r.response + string(t)

				// fmt.Println(r.response)
				// fmt.Println("current llm resp", string(t))
				// fmt.Println("current r.response", r.response)
			}

		}
	}
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

		return r.Streaming()
	}

	return r, nil
}

func (r llm_response) View() string {
	out, err := glamour.Render(r.response, "dark")
	_ = err
	style := lipgloss.NewStyle()
	return style.Border(lipgloss.NormalBorder()).Render(out)
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
		Width(w).Align(lipgloss.Center)

	quote := "Curiosity changes everything."

	fmt.Println(style.Render(quote))
}

func ask() {
	var question string

	huh.NewInput().
		Title("Ask anything...").
		Value(&question).
		// WithTheme()
		Run() // this is blocking...

	out, err := glamour.Render("# "+question, "dark")

	_ = err
	fmt.Println(out)

	p := tea.NewProgram(initialLLMResponse(question))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
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

	}

	print_banner()
	ask()
}
