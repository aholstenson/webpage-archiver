package progress

import (
	"os"
	"strconv"
	"time"

	"github.com/aholstenson/webpage-archiver/pkg/network"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var quitKey = key.NewBinding(
	key.WithKeys("ctrl+c"),
)

var styleTime = lipgloss.NewStyle().
	Width(5)

var styleCurrentURL = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1)

var styleMethod = lipgloss.NewStyle().
	Bold(true).
	PaddingRight(1)

var styleURL = lipgloss.NewStyle().Faint(true)

var styleDefault = lipgloss.NewStyle().
	Bold(true).
	PaddingRight(1)

var style2xx = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#04B575")).
	PaddingRight(1)

var style3xx = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FDD835")).
	PaddingRight(1)

var style4xx = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFA726")).
	PaddingRight(1)

var style5xx = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FF7043")).
	PaddingRight(1)

type interactiveReporter struct {
	ctxCancel func()

	logMessages        []any
	logMessagesChannel chan any

	url        string
	urlChannel chan string

	program   *tea.Program
	stopwatch stopwatch.Model
	spinner   spinner.Model
}

func NewInteractiveReporter(cancel func()) (Reporter, error) {
	m := &interactiveReporter{
		ctxCancel: cancel,

		logMessages:        make([]any, 0),
		logMessagesChannel: make(chan any),

		urlChannel: make(chan string),

		stopwatch: stopwatch.NewWithInterval(time.Millisecond * 100),
		spinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
	}

	p := tea.NewProgram(m)
	m.program = p
	go func() {
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}

		cancel()
	}()

	return m, nil
}

func (m *interactiveReporter) Close() error {
	m.program.Quit()
	return nil
}

func (c *interactiveReporter) Start(url string) {
	c.urlChannel <- url
}

func (m *interactiveReporter) Debug(msg string) {
	m.logMessagesChannel <- debugMessage(msg)
}

func (m *interactiveReporter) Error(err error, msg string) {
	m.logMessagesChannel <- errorMessage(msg + ": " + err.Error())
}

func (m *interactiveReporter) Request(req *network.Request) {
	m.logMessagesChannel <- req
}

func (m *interactiveReporter) Response(res *network.Response) {
	m.logMessagesChannel <- res
}

func (m *interactiveReporter) Init() tea.Cmd {
	return tea.Batch(
		m.waitForLogMessage(m.logMessagesChannel),
		m.waitForURL(m.urlChannel),
		m.stopwatch.Init(),
		m.spinner.Tick,
	)
}

func (m *interactiveReporter) View() string {
	s := m.spinner.View() + styleTime.Render(m.stopwatch.View()) + styleCurrentURL.Render(m.url) + "\n\n"
	if len(m.logMessages) > 0 {
		for _, m := range m.logMessages {
			switch msg := m.(type) {
			case *network.Request:
				s += styleMethod.Render(msg.Method) + styleURL.Render(msg.URL) + "\n"
			case *network.Response:
				statusStyle := styleDefault
				if msg.StatusCode >= 200 && msg.StatusCode < 300 {
					statusStyle = style2xx
				} else if msg.StatusCode >= 300 && msg.StatusCode < 400 {
					statusStyle = style3xx
				} else if msg.StatusCode >= 400 && msg.StatusCode < 500 {
					statusStyle = style4xx
				} else if msg.StatusCode >= 500 && msg.StatusCode < 600 {
					statusStyle = style5xx
				}
				s += statusStyle.Render(strconv.Itoa(msg.StatusCode)+" "+msg.StatusPhrase) + styleURL.Render(msg.URL) + "\n"
			case debugMessage:
				s += string(msg) + "\n"
			case errorMessage:
				s += "❌ " + string(msg) + "\n"
			}
		}
	}
	return s
}

func (m *interactiveReporter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, quitKey):
			return m, tea.Quit
		}
	case url:
		m.url = string(msg)
		return m, m.waitForURL(m.urlChannel)
	case activityMsg:
		if len(m.logMessages) >= 10 {
			m.logMessages = m.logMessages[1:10]
		}
		m.logMessages = append(m.logMessages, msg.data)
		return m, m.waitForLogMessage(m.logMessagesChannel)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case stopwatch.TickMsg, stopwatch.StartStopMsg:
		var cmd tea.Cmd
		m.stopwatch, cmd = m.stopwatch.Update(msg)
		return m, cmd
	default:
		return m, nil
	}

	return m, nil
}

func (m *interactiveReporter) waitForLogMessage(c chan any) tea.Cmd {
	return func() tea.Msg {
		return activityMsg(activityMsg{
			data: <-c,
		})
	}
}

func (m *interactiveReporter) waitForURL(c chan string) tea.Cmd {
	return func() tea.Msg {
		return url(<-c)
	}
}

type activityMsg struct {
	data any
}

type url string

type debugMessage string

type errorMessage string
