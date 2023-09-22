package models_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xor111xor/pomodoro-go/internal/models"
	"github.com/xor111xor/pomodoro-go/internal/repository"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		name   string
		input  [3]time.Duration
		expect models.IntervalConfig
	}{
		{
			name: "Default",
			expect: models.IntervalConfig{
				PomoDuration:       25 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
			},
		},
		{
			name: "Single parametrs",
			input: [3]time.Duration{
				20 * time.Minute,
			},
			expect: models.IntervalConfig{
				PomoDuration:       20 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
			},
		},
		{
			name: "Multi parametrs",
			input: [3]time.Duration{
				20 * time.Minute,
				12 * time.Minute,
				7 * time.Minute,
			},
			expect: models.IntervalConfig{
				PomoDuration:       20 * time.Minute,
				LongBreakDuration:  12 * time.Minute,
				ShortBreakDuration: 7 * time.Minute,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var repo models.Repository
			config, _ := models.NewConfig(
				repo,
				tc.input[0],
				tc.input[1],
				tc.input[2],
			)
			if config.PomoDuration != tc.expect.PomoDuration {
				t.Errorf("Expected Duration %q got %q instead",
					tc.expect.PomoDuration, config.PomoDuration)
			}
			if config.LongBreakDuration != tc.expect.LongBreakDuration {
				t.Errorf("Expected Duration %q got %q instead",
					tc.expect.LongBreakDuration, config.LongBreakDuration)
			}
			if config.ShortBreakDuration != tc.expect.ShortBreakDuration {
				t.Errorf("Expected Duration %q got %q instead\n",
					tc.expect.ShortBreakDuration, config.ShortBreakDuration)
			}
		})
	}
}
func getRepo(t *testing.T) (models.Repository, func()) {
	t.Helper()
	return repository.NewInMemoryRepo(), func() {}
}

func TestGetInterval(t *testing.T) {
	repo, cleanup := getRepo(t)
	defer cleanup()

	const duration = 1 * time.Millisecond
	config, _ := models.NewConfig(repo, 3*duration, 2*duration, duration)

	for i := 1; i <= 16; i++ {
		var (
			expCategory string
			expDuration time.Duration
		)

		switch {
		case i%2 != 0:
			expDuration = 3 * duration
			expCategory = models.PomodoCategory
		case i%8 == 0:
			expDuration = 2 * duration
			expCategory = models.LongBreakCategory
		case i%2 == 0:
			expDuration = duration
			expCategory = models.ShortBreakCategory
		}

		tesName := fmt.Sprintf("%s %d", expCategory, i)

		t.Run(tesName, func(t *testing.T) {
			res, err := models.GetInterval(config)
			if err != nil {
				t.Errorf("Expected no error, got %q.\n", err)
			}
			noop := func(models.Interval) {}

			if err := res.Start(context.Background(), config, noop, noop, noop); err != nil {
				t.Fatal(err)
			}

			if res.Category != expCategory {
				t.Errorf("Expected category %q, got %q.\n", expCategory, res.Category)
			}
			if res.TimePlanning != expDuration {
				t.Errorf("Expected TimePlanning %q, got %q.\n", expDuration, res.TimePlanning)
			}
			if res.State != models.StateNotStarted {
				t.Errorf("Expected State %q, got %q.\n", models.StateNotStarted, res.State)
			}

			ui, err := repo.ByID(res.ID)
			if err != nil {
				t.Errorf("Expected no error, got %q.\n", err)
			}
			if ui.State != models.StateDone {
				t.Errorf("Expected state %q, got %q.\n", models.StateDone, ui.State)
			}
		})

	}

}
