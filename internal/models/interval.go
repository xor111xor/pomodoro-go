package models

import "fmt"

const (
	PomodoCategory     = "Pomodoro"
	LongBreakCategory  = "Long"
	ShortBreakCategory = "Short"
)

const (
	StateNotStarted = iota
	StateRunning
	StatePaused
	StateCanceled
	StateDone
)

var (
	ErrNoInterval = fmt.Errorf("No interval")
)
