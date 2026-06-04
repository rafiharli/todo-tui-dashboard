package tui

import (
	"fmt"
	"strings"
	"time"
	"todo-dashboard/internal/db"
	"todo-dashboard/internal/models"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type inputStep int

const (
	stepTitle inputStep = iota
	stepDate
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
	inputStep   inputStep
	tempTitle   string
	textInput   textinput.Model
	progress    progress.Model
	help        help.Model
	quitting    bool
	width       int
	height      int
	lastError   error
	now         time.Time
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
		now:       time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchTasks(), tick())
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
	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()

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
				if m.inputStep == stepTitle {
					if m.textInput.Value() != "" {
						m.tempTitle = m.textInput.Value()
						m.textInput.SetValue("")
						m.textInput.Placeholder = "Due date? (e.g. 2006-01-02 15:04 or leave empty)"
						m.inputStep = stepDate
						return m, nil
					}
				} else {
					var dueDate *time.Time
					dateVal := m.textInput.Value()
					if dateVal != "" {
						// Simple parsing logic
						formats := []string{
							"2006-01-02 15:04",
							"2006-01-02",
							"02-01-2006 15:04",
							"02-01-2006",
							"15:04",
						}
						for _, f := range formats {
							t, err := time.ParseInLocation(f, dateVal, time.Local)
							if err == nil {
								// If only time was provided, assume today
								if f == "15:04" {
									now := time.Now()
									t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
								}
								dueDate = &t
								break
							}
						}
						if dueDate == nil {
							m.lastError = fmt.Errorf("invalid date format")
							return m, nil
						}
					}

					err := m.db.AddTask(m.tempTitle, models.Medium, dueDate)
					if err != nil {
						m.lastError = err
					}
					m.textInput.SetValue("")
					m.textInput.Placeholder = "What needs to be done?"
					m.inputMode = false
					m.inputStep = stepTitle
					m.lastError = nil
					return m, m.fetchTasks()
				}
			case msg.Type == tea.KeyEsc:
				m.inputMode = false
				m.inputStep = stepTitle
				m.textInput.SetValue("")
				m.textInput.Placeholder = "What needs to be done?"
				m.lastError = nil
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
			m.inputStep = stepTitle
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
			Padding(0, 1)

	clockStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			MarginBottom(1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)

	doneStyle = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("#626262"))

	todoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))

	dueStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
	overdueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Bold(true)

	priorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	priorityMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FDFFAD"))
	priorityLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#87D787"))

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(60)
)

func (m Model) View() string {
	if m.quitting {
		return "See you tomorrow!\n"
	}

	var s strings.Builder

	// Header construction
	headerTitle := titleStyle.Render(" DAILY DASHBOARD ")
	clockStr := clockStyle.Render(m.now.Format("15:04:05"))
	dateStr := dateStyle.Render(m.now.Format("Monday, 02 Jan 2006"))

	header := lipgloss.JoinHorizontal(lipgloss.Center, headerTitle, dateStr, clockStr)
	s.WriteString(headerStyle.Width(m.width - 4).Render(header))
	s.WriteString("\n")

	// Main content in a box
	var content strings.Builder

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

	content.WriteString(fmt.Sprintf("Progress: %d/%d tasks completed\n", doneCount, total))
	content.WriteString(m.progress.ViewAs(pct))
	content.WriteString("\n\n")

	// Task list
	if m.inputMode {
		prompt := "TASK TITLE"
		if m.inputStep == stepDate {
			prompt = "DUE DATE (Optional)"
		}
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Render(prompt) + "\n")
		content.WriteString(m.textInput.View() + "\n\n")
		if m.inputStep == stepDate {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("Formats: YYYY-MM-DD, HH:MM, or DD-MM-YYYY") + "\n")
		}
	} else {
		if len(m.tasks) == 0 {
			content.WriteString("No tasks yet. Press 'n' to add one.\n\n")
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

				dueStr := ""
				if task.DueDate != nil {
					isOverdue := task.DueDate.Before(m.now) && !task.Status
					style := dueStyle
					if isOverdue {
						style = overdueStyle
					}
					
					// Nice formatting for due date
					format := "02 Jan 15:04"
					if task.DueDate.Year() != m.now.Year() {
						format = "02 Jan 2006"
					}
					dueStr = style.Render(" (Due: " + task.DueDate.Format(format) + ")")
				}

				content.WriteString(fmt.Sprintf("%s %s %s%s\n", cursor, checked, taskTitle, dueStr))
			}
			content.WriteString("\n")
		}
	}

	if m.lastError != nil {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("Error: %v", m.lastError)))
		content.WriteString("\n")
	}

	s.WriteString(boxStyle.Width(m.width - 6).Render(content.String()))
	s.WriteString("\n\n")

	s.WriteString(m.help.View(keys))

	return docStyle.Render(s.String())
}
