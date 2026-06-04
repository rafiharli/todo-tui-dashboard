package db

import (
	"database/sql"
	"time"
	"todo-dashboard/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		priority INTEGER DEFAULT 0,
		status BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		due_date DATETIME
	);`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Migration: Add due_date column if it doesn't exist
	_, _ = db.Exec("ALTER TABLE tasks ADD COLUMN due_date DATETIME")

	return &DB{conn: db}, nil
}

func (d *DB) AddTask(title string, priority models.Priority, dueDate *time.Time) error {
	_, err := d.conn.Exec("INSERT INTO tasks (title, priority, due_date) VALUES (?, ?, ?)", title, int(priority), dueDate)
	return err
}

func (d *DB) GetTasks() ([]models.Task, error) {
	rows, err := d.conn.Query("SELECT id, title, priority, status, created_at, due_date FROM tasks ORDER BY status ASC, priority DESC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var p int
		err := rows.Scan(&t.ID, &t.Title, &p, &t.Status, &t.CreatedAt, &t.DueDate)
		if err != nil {
			return nil, err
		}
		t.Priority = models.Priority(p)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (d *DB) ToggleTask(id int) error {
	_, err := d.conn.Exec("UPDATE tasks SET status = NOT status WHERE id = ?", id)
	return err
}

func (d *DB) DeleteTask(id int) error {
	_, err := d.conn.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}
