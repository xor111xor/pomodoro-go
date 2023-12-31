// go:build inmemory

package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xor111xor/pomodoro-go/internal/models"
)

type InMemoryRepo struct {
	sync.RWMutex
	intervals []models.Interval
}

// Create new in-memory repository
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		intervals: []models.Interval{},
	}
}

func (in *InMemoryRepo) Create(i models.Interval) (int64, error) {
	in.Lock()
	defer in.Unlock()

	i.ID = int64(len(in.intervals) + 1)

	in.intervals = append(in.intervals, i)
	return i.ID, nil
}

func (in *InMemoryRepo) Update(i models.Interval) error {
	in.Lock()
	defer in.Unlock()

	if i.ID == 0 {
		return fmt.Errorf("%w: %d", models.ErrInvalidID, i.ID)
	}

	in.intervals[i.ID-1] = i
	return nil
}

func (in *InMemoryRepo) Last() (models.Interval, error) {
	in.RLock()
	defer in.RUnlock()

	i := models.Interval{}
	if len(in.intervals) == 0 {
		return i, models.ErrNoIntervals
	}

	last := in.intervals[len(in.intervals)-1]
	return last, nil
}

func (in *InMemoryRepo) ByID(id int64) (models.Interval, error) {
	in.RLock()
	defer in.RUnlock()

	i := models.Interval{}
	if id == 0 {
		return i, fmt.Errorf("%w: %d", models.ErrInvalidID, id)
	}

	i = in.intervals[id-1]
	return i, nil
}

func (in *InMemoryRepo) Breaks(count int) ([]models.Interval, error) {
	breaks := []models.Interval{}

	for i := len(in.intervals) - 1; i >= 0; i-- {
		if in.intervals[i].Category != models.PomodoCategory {
			breaks = append(breaks, in.intervals[i])
		}
		if len(breaks) == count {
			return breaks, nil
		}
	}
	return breaks, nil
}
func (in *InMemoryRepo) CategorySummary(day time.Time, filter string) (time.Duration, error) {
	// Return daily summary
	in.RLock()
	defer in.RUnlock()

	var d time.Duration
	filter = strings.Trim(filter, "%")

	for _, i := range in.intervals {
		if i.TimeStart.Year() == day.Year() &&
			i.TimeStart.YearDay() == day.YearDay() {
			if strings.Contains(i.Category, filter) {
				d += i.TimeActual
			}
		}
	}
	return d, nil
}
