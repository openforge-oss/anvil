package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/openforge-oss/anvil/internal/driver"
	"github.com/openforge-oss/anvil/internal/pipeline"
)

const tailLines = 8

type eventMsg pipeline.Event
type closedMsg struct{}

func waitEvent(ch <-chan pipeline.Event) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-ch
		if !ok {
			return closedMsg{}
		}
		return eventMsg(e)
	}
}

type model struct {
	events  <-chan pipeline.Event
	items   []pipeline.Item
	status  []driver.Status
	spinner spinner.Model
	tail    []string
	success bool
}

func newModel(ch <-chan pipeline.Event) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return model{events: ch, spinner: s}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitEvent(m.events))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case closedMsg:
		return m, tea.Quit
	case eventMsg:
		m.apply(pipeline.Event(msg))
		return m, waitEvent(m.events)
	}
	return m, nil
}

func (m *model) apply(e pipeline.Event) {
	switch e.Kind {
	case pipeline.KindPlan:
		m.items = e.Items
		m.status = make([]driver.Status, len(e.Items))
	case pipeline.KindStepStart:
		if e.Index < len(m.status) {
			m.status[e.Index] = driver.Running
		}
		m.tail = nil
	case pipeline.KindLine:
		m.tail = append(m.tail, e.Line)
		if len(m.tail) > tailLines {
			m.tail = m.tail[len(m.tail)-tailLines:]
		}
	case pipeline.KindStepDone:
		if e.Index < len(m.status) {
			m.status[e.Index] = e.Status
		}
	case pipeline.KindDone:
		m.success = e.Success
	}
}

var (
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	runStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle  = lipgloss.NewStyle().Faint(true)
)

func (m model) View() string {
	var b strings.Builder
	for i, it := range m.items {
		st := driver.Pending
		if i < len(m.status) {
			st = m.status[i]
		}
		marker, label := m.row(st, it)
		fmt.Fprintf(&b, "%s %s\n", marker, label)
	}
	if len(m.tail) > 0 {
		b.WriteString("\n")
		for _, line := range m.tail {
			b.WriteString(dimStyle.Render("  "+line) + "\n")
		}
	}
	return b.String()
}

func (m model) row(st driver.Status, it pipeline.Item) (marker, label string) {
	name := fmt.Sprintf("%s: %s", it.Phase, it.Name)
	switch st {
	case driver.Running:
		return runStyle.Render(m.spinner.View()), name
	case driver.OK:
		return okStyle.Render("[ok]"), name
	case driver.Failed:
		return failStyle.Render("[!!]"), name
	case driver.Skipped:
		return dimStyle.Render("[--]"), dimStyle.Render(name + " (skipped)")
	default:
		return dimStyle.Render("[ ]"), dimStyle.Render(name)
	}
}

// Run renders pipeline events with an interactive TUI and returns success.
func Run(events <-chan pipeline.Event) (bool, error) {
	final, err := tea.NewProgram(newModel(events)).Run()
	if err != nil {
		return false, err
	}
	return final.(model).success, nil
}
