package app

import (
	"context"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
)

type widgets struct {
	donTimer       *donut.Donut
	disType        *segmentdisplay.SegmentDisplay
	txtInfo        *text.Text
	txtTimer       *text.Text
	updateDonTimer chan []int
	updateTxtInfo  chan string
	updateTxtTimer chan string
	updateTxtType  chan string
}

func (w *widgets) update(timer []int, txtInfo, txtTimer, txtType string, redrawCh chan<- bool) {
	if len(timer) > 0 {
		w.updateDonTimer <- timer
	}
	if txtInfo != "" {
		w.updateTxtInfo <- txtInfo
	}
	if txtTimer != "" {
		w.updateTxtTimer <- txtTimer
	}
	if txtType != "" {
		w.updateTxtType <- txtType
	}
	redrawCh <- true
}

func newWidgets(ctx context.Context, errorCh chan<- error) (*widgets, error) {
	w := &widgets{}
	var err error

	w.updateDonTimer = make(chan []int)
	w.updateTxtInfo = make(chan string)
	w.updateTxtTimer = make(chan string)
	w.updateTxtType = make(chan string)

	w.donTimer, err := newDonut(ctx, w.updateDonTimer, errorCh)
	if err != nil {
		return nil, err
	}

	w.disType, err := newSegmentDisplay(ctx, w.updateTxtType, errorCh)
	if err != nil {
		return nil, err
	}
	w.txtInfo, err := newText(ctx, w.updateTxtInfo, errorCh)
	if err != nil {
		return nil, err
	}
	w.txtTimer, err := newText(txt, w.updateTxtTimer, errorCh)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func newText(ctx context.Context, updateText <-chan string, errorCh chan<- error) (*text.Text, error) {
	txt, err := text.New()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case t := <-updateText:
				txt.Reset()
				errorCh <- txt.Write(t)
			case <-ctx.Done():
				return
			}
		}
	}()
	return txt, nil
}
