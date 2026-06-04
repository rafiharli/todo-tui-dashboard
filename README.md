# 🚀 Advanced TUI Daily To-Do Dashboard

A sophisticated, keyboard-centric Daily To-Do Dashboard built with **Go** and the **Charm (Bubble Tea)** ecosystem. Designed for developers who live in the terminal and want a fast, aesthetic way to manage their daily tasks.

<img width="1175" height="581" alt="Screenshot 2026-06-04 150621" src="https://github.com/user-attachments/assets/b7f98835-4109-4636-9a67-70eee7475cb1" />

## ✨ Features

- **Aesthetic Interface:** Styled with `lipgloss` for a modern, colorful terminal experience.
- **Due Dates & Times:** Set deadlines for your tasks to stay on track.
- **Overdue Alerts:** Visual indicators for tasks that have passed their deadline.
- **Progress Tracking:** Real-time progress bar that visualizes your daily completion rate.
- **Keyboard-Centric:** Vim-inspired navigation (`j/k`) and intuitive shortcuts.
- **Persistent Storage:** Built-in SQLite integration to ensure your tasks are saved locally.
- **Responsive Design:** Gracefully handles terminal resizing.
- **Clean Architecture:** Modular Go code using the Model-View-Update (MVU) pattern.

## 🛠️ Tech Stack

- **Language:** [Go](https://go.dev/)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **UI Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Database:** [SQLite](https://sqlite.org/) (via `go-sqlite3`)

## 🚀 Getting Started

### Prerequisites

- Go 1.24 or higher installed on your system.

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/todo-dashboard.git
   cd todo-dashboard
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run it!**
   ```bash
   go run main.go
   ```

## ⌨️ Keybindings

| Key | Action |
|-----|--------|
| `n` | Add a new task (Step 1: Title, Step 2: Optional Date) |
| `Space` | Toggle task completion (Done/Todo) |
| `↑/k` | Move cursor up |
| `↓/j` | Move cursor down |
| `x` | Delete selected task |
| `q` / `Ctrl+C` | Quit application |
| `Esc` | Cancel adding new task |
| `Enter` | Confirm input |

### Date Formats Supported
When adding a task, you can use several formats for the due date:
- `YYYY-MM-DD HH:MM` (e.g., `2026-06-05 14:00`)
- `YYYY-MM-DD`
- `DD-MM-YYYY`
- `HH:MM` (Assumes today)

## 📁 Project Structure

```text
.
├── main.go             # Entry point
├── internal/
│   ├── db/             # SQLite database layer
│   ├── models/         # Task data models
│   └── tui/            # Bubble Tea models and view logic
└── go.mod              # Go dependencies
```

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

---
Built with ❤️ by Rafi Harlianto
