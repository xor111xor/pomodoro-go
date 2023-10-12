package app

import (
	"context"
	"fmt"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/xor111xor/pomodoro-go/internal/models"
)

type buttons struct {
	btStart *button.Button
	btPause *button.Button
}

func newButtons(ctx context.Context, config *models.IntervalConfig, w *widgets, s *summary, redrawCh chan<- bool, errorCh chan<- error) (*buttons, error) {
	startInterval := func() {
		i, err := models.GetInterval(config)
		errorCh <- err
		start := func(i models.Interval) {
			message := "Take a brake"
			if i.Category == models.PomodoCategory {
				message = "Focus on your task"
			}
			w.update([]int{}, message, "", i.Category, redrawCh)
		}
		end := func(models.Interval) {
			w.update([]int{}, "", "Nothing running", "", redrawCh)
			s.update(redrawCh)
		}
		periodic := func(i models.Interval) {
			w.update(
				[]int{int(i.TimeActual), int(i.TimePlanning)},
				"",
				fmt.Sprint(i.TimePlanning-i.TimeActual),
				"",
				redrawCh,
			)
		}
		errorCh <- i.Start(ctx, config, start, periodic, end)
	}
	pauseInterval := func() {
		i, err := models.GetInterval(config)
		if err != nil {
			errorCh <- err
		}

		if err := i.Pause(config); err != nil {
			if err == models.ErrIntervalNotRunning {
				return
			}
			errorCh <- err
			return
		}
		w.update([]int{}, "Paused, press start to continue...", "", "", redrawCh)
	}

	btStart, err := button.New("(s)tart", func() error {
		go startInterval()
		return nil
	},
		button.GlobalKey('s'),
		button.WidthFor("(p)ause"),
		button.Height(2),
	)
	if err != nil {
		return nil, err
	}

	btPause, err := button.New("(p)ause", func() error {
		go pauseInterval()
		return nil
	},
		button.FocusedFillColor(cell.ColorAqua),
		button.GlobalKey('p'),
		button.Height(2),
	)
	if err != nil {
		return nil, err
	}

	return &buttons{btStart: btStart, btPause: btPause}, nil

}
