package main

import (
	"fmt"
	"os"
	"path/filepath"
	"todo-dashboard/internal/db"
	"todo-dashboard/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Setup database path
	home, _ := os.UserHomeDir()
	dbDir := filepath.Join(home, ".todo-dashboard")
	os.MkdirAll(dbDir, 0755)
	dbPath := filepath.Join(dbDir, "tasks.db")

	database, err := db.NewDB(dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	p := tea.NewProgram(tui.NewModel(database), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
