package tui

import (
	"fmt"
	"strings"
	"todo-dashboard/internal/db"
	"todo-dashboard/internal/models"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	New    key.Binding
	Delete key.Binding
	Quit   key.Binding
	Enter  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.New, k.Delete, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.New, k.Delete, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new task"),
	),
	Delete: key.NewBinding(
		key.WithKeys("x", "delete"),
		key.WithHelp("x", "delete"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
}

type Model struct {
	db          *db.DB
	tasks       []models.Task
	cursor      int
	inputMode   bool
	textInput   textinput.Model
	progress    progress.Model
	help        help.Model
	quitting    bool
	width       int
	height      int
	lastError   error
}

func NewModel(database *db.DB) Model {
	ti := textinput.New()
	ti.Placeholder = "What needs to be done?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	p := progress.New(progress.WithDefaultGradient())

	return Model{
		db:        database,
		textInput: ti,
		progress:  p,
		help:      help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchTasks()
}

func (m Model) fetchTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.db.GetTasks()
		if err != nil {
			return err
		}
		return tasks
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		return m, nil

	case []models.Task:
		m.tasks = msg
		if m.cursor >= len(m.tasks) {
			m.cursor = len(m.tasks) - 1
		}
		if m.cursor < 0 && len(m.tasks) > 0 {
			m.cursor = 0
		}
		return m, nil

	case error:
		m.lastError = msg
		return m, nil

	case tea.KeyMsg:
		if m.inputMode {
			switch {
			case key.Matches(msg, keys.Enter):
				if m.textInput.Value() != "" {
					err := m.db.AddTask(m.textInput.Value(), models.Medium)
					if err != nil {
						m.lastError = err
					}
					m.textInput.SetValue("")
					m.inputMode = false
					return m, m.fetchTasks()
				}
			case msg.Type == tea.KeyEsc:
				m.inputMode = false
				m.textInput.SetValue("")
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Toggle):
			if len(m.tasks) > 0 {
				err := m.db.ToggleTask(m.tasks[m.cursor].ID)
				if err != nil {
					m.lastError = err
				}
				return m, m.fetchTasks()
			}

		case key.Matches(msg, keys.New):
			m.inputMode = true
			m.textInput.Focus()
			return m, nil

		case key.Matches(msg, keys.Delete):
			if len(m.tasks) > 0 {
				err := m.db.DeleteTask(m.tasks[m.cursor].ID)
				if err != nil {
					m.lastError = err
				}
				return m, m.fetchTasks()
			}
		}
	}

	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)

	doneStyle = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("#626262"))

	todoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))

	priorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	priorityMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FDFFAD"))
	priorityLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#87D787"))
)

func (m Model) View() string {
	if m.quitting {
		return "See you tomorrow!\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(" DAILY DASHBOARD "))
	s.WriteString("\n\n")

	// Progress section
	doneCount := 0
	for _, t := range m.tasks {
		if t.Status {
			doneCount++
		}
	}
	
	total := len(m.tasks)
	pct := 0.0
	if total > 0 {
		pct = float64(doneCount) / float64(total)
	}

	s.WriteString(fmt.Sprintf("Progress: %d/%d tasks completed\n", doneCount, total))
	s.WriteString(m.progress.ViewAs(pct))
	s.WriteString("\n\n")

	// Task list
	if m.inputMode {
		s.WriteString("New Task: " + m.textInput.View() + "\n\n")
	} else {
		if len(m.tasks) == 0 {
			s.WriteString("No tasks yet. Press 'n' to add one.\n\n")
		} else {
			for i, task := range m.tasks {
				cursor := " "
				if m.cursor == i {
					cursor = cursorStyle.Render(">")
				}

				checked := "[ ]"
				taskTitle := todoStyle.Render(task.Title)
				if task.Status {
					checked = "[x]"
					taskTitle = doneStyle.Render(task.Title)
				}

				s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checked, taskTitle))
			}
			s.WriteString("\n")
		}
	}

	if m.lastError != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("Error: %v", m.lastError)))
		s.WriteString("\n")
	}

	s.WriteString(m.help.View(keys))

	return docStyle.Render(s.String())
}
