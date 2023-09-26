package models

import (
	"context"
	"fmt"
	"time"
)

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
	ErrNoIntervals        = fmt.Errorf("No interval")
	ErrIntervalNotRunning = fmt.Errorf("Interval not running")
	ErrIntervalCompleted  = fmt.Errorf("Interval completed")
	ErrInvalidState       = fmt.Errorf("Intervarl invalid state")
	ErrInvalidID          = fmt.Errorf("Interval invalid id")
)

type Interval struct {
	ID           int
	Category     string
	State        int
	TimeStart    time.Time
	TimePlanning time.Duration
	TimeActual   time.Duration
}

type Repository interface {
	Create(i Interval) (int, error)
	Update(i Interval) error
	Last() (Interval, error)
	ByID(int) (Interval, error)
	Breaks(int) ([]Interval, error)
}

type IntervalConfig struct {
	Repo               Repository
	PomoDuration       time.Duration
	LongBreakDuration  time.Duration
	ShortBreakDuration time.Duration
}

// Init new config
func NewConfig(repo Repository, pomo, long, short time.Duration) (*IntervalConfig, error) {
	config := &IntervalConfig{
		Repo:               repo,
		PomoDuration:       25 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
	}

	if pomo > 0 {
		config.PomoDuration = pomo
	}
	if long > 0 {
		config.LongBreakDuration = long
	}
	if short > 0 {
		config.ShortBreakDuration = short
	}

	return config, nil
}

// Recognize next category
func NextCategory(r Repository) (string, error) {
	last, err := r.Last()
	if err != nil && err == ErrNoIntervals {
		return PomodoCategory, nil
	}
	if err != nil {
		return "", err
	}

	if last.Category == LongBreakCategory || last.Category == ShortBreakCategory {
		return PomodoCategory, nil
	}

	breaks, err := r.Breaks(3)
	if err != nil {
		return "", nil
	}

	if len(breaks) < 3 {
		return ShortBreakCategory, nil
	}

	for _, i := range breaks {
		if i.Category == LongBreakCategory {
			return ShortBreakCategory, nil
		}
	}

	return LongBreakCategory, nil
}

// Return new interval
func NewInterval(config *IntervalConfig) (Interval, error) {
	i := Interval{}

	category, err := NextCategory(config.Repo)
	if err != nil {
		return i, nil
	}

	i.Category = category

	switch category {
	case PomodoCategory:
		i.TimePlanning = config.PomoDuration
	case LongBreakCategory:
		i.TimePlanning = config.LongBreakDuration
	case ShortBreakCategory:
		i.TimePlanning = config.ShortBreakDuration
	}

	if i.ID, err = config.Repo.Create(i); err != nil {
		return i, err
	}
	return i, nil
}

// Return running, stoped or new interval
func GetInterval(config *IntervalConfig) (Interval, error) {
	i := Interval{}

	i, err := config.Repo.Last()
	if err != nil && err != ErrNoIntervals {
		return i, err
	}

	if err == nil && i.State != StateCanceled && i.State != StateDone {
		return i, nil
	}

	return NewInterval(config)
}

type Callback func(Interval)

// Performing action for interval
func tick(ctx context.Context, config *IntervalConfig, start, periodic, end Callback, id int) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	i, err := config.Repo.ByID(id)
	if err != nil {
		return err
	}

	expire := time.After(i.TimePlanning - i.TimeActual)

	start(i)

	for {
		select {
		case <-ticker.C:
			i, err := config.Repo.ByID(id)
			if err != nil {
				return err
			}

			if i.State == StatePaused {
				return nil
			}

			i.TimeActual += time.Second
			if err := config.Repo.Update(i); err != nil {
				return err
			}
			periodic(i)
		case <-expire:
			i, err := config.Repo.ByID(id)
			if err != nil {
				return err
			}
			i.State = StateDone
			end(i)
			return config.Repo.Update(i)
		case <-ctx.Done():
			i, err := config.Repo.ByID(id)
			if err != nil {
				return err
			}

			i.State = StateCanceled
			return config.Repo.Update(i)
		}
	}
}

func (i Interval) Start(ctx context.Context, config *IntervalConfig, start, periodic, end Callback) error {
	switch i.State {
	case StateRunning:
		return nil
	case StateNotStarted:
		i.TimeStart = time.Now()
		fallthrough
	case StatePaused:
		i.State = StateRunning
		if err := config.Repo.Update(i); err != nil {
			return err
		}
		return tick(ctx, config, start, periodic, end, i.ID)
	case StateCanceled, StateDone:
		return fmt.Errorf("%w: Cannot Start", ErrIntervalCompleted)
	default:
		return fmt.Errorf("%w: %d", ErrInvalidState, i.State)
	}
}

func (i Interval) Pause(config *IntervalConfig) error {
	if i.State != StateRunning {
		return ErrIntervalNotRunning
	}
	i.State = StatePaused
	return config.Repo.Update(i)
}
