package models_test

import (
	"context"
	"errors"
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

func TestPause(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	config, err := models.NewConfig(repo, duration, duration, duration)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name        string
		start       bool
		expState    int
		expDuration time.Duration
	}{
		{
			name:        "NotStarted",
			start:       false,
			expState:    models.StateNotStarted,
			expDuration: 0,
		},
		{
			name:        "Paused",
			start:       true,
			expState:    models.StatePaused,
			expDuration: duration / 2,
		},
	}
	expError := models.ErrIntervalNotRunning

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			i, err := models.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			start := func(models.Interval) {}
			end := func(models.Interval) {
				t.Errorf("End callback should not be executed")
			}
			periodic := func(i models.Interval) {
				if err := i.Pause(config); err != nil {
					t.Fatal(err)
				}
			}

			if tc.start {
				if err := i.Start(ctx, config, start, periodic, end); err != nil {
					t.Fatal(err)
				}
			}

			i, err = models.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			err = i.Pause(config)
			if err != nil {
				if !errors.Is(err, expError) {
					t.Fatalf("Expected error %q got %q", expError, err)
				}
			}
			if err == nil {
				t.Fatalf("Expected error %q, got nil", expError)
			}

			i, err = repo.ByID(i.ID)
			if err != nil {
				t.Fatal(err)
			}

			if i.State != tc.expState {
				t.Errorf("Expected state %q, got %q.\n", tc.expState, i.State)
			}

			if i.TimeActual != tc.expDuration {
				t.Errorf("Expected duration %q, got %q.\n", tc.expDuration, i.TimeActual)
			}
			cancel()
		})
	}

}

func TestStart(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	config, err := models.NewConfig(repo, duration, duration, duration)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name        string
		cancel      bool
		expState    int
		expDuration time.Duration
	}{
		{
			name:        "Finish",
			cancel:      false,
			expState:    models.StateDone,
			expDuration: duration,
		},
		{
			name:        "Cancel",
			cancel:      true,
			expState:    models.StateCanceled,
			expDuration: duration / 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			i, err := models.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			start := func(i models.Interval) {
				if i.State != models.StateRunning {
					t.Errorf("Expected state %d, got %d", models.StateRunning, tc.expState)
				}

				if i.TimeActual >= i.TimePlanning {
					t.Errorf("Expected ActualDuration %q, less than Planned %q.\n", i.TimeActual, i.TimePlanning)
				}
			}

			end := func(i models.Interval) {
				if i.State != tc.expState {
					t.Errorf("Expected state %d, got %d", i.State, tc.expState)
				}
				if tc.cancel {
					t.Error("End callback should not be executed")
				}
			}

			periodic := func(i models.Interval) {
				if i.State != models.StateRunning {
					t.Fatalf("Expected state %q, got %q", models.StateRunning, i.State)
				}

				if tc.cancel {
					cancel()
				}
			}

			if err := i.Start(ctx, config, start, periodic, end); err != nil {
				t.Fatal(err)
			}

			i, err = repo.ByID(i.ID)
			if err != nil {
				t.Fatal(err)
			}

			if tc.expState != i.State {
				t.Errorf("Expected state %q, got %q", tc.expState, i.State)
			}
			if tc.expDuration != i.TimeActual {
				t.Errorf("Expected duration %d, got %d", tc.expDuration, i.TimeActual)
			}

			cancel()
		})
	}
}
