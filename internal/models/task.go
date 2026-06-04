package models

import "time"

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

func (p Priority) String() string {
	return [...]string{"Low", "Medium", "High"}[p]
}

type Task struct {
	ID        int
	Title     string
	Priority  Priority
	Status    bool // false = Todo, true = Done
	CreatedAt time.Time
	DueDate   *time.Time
}
