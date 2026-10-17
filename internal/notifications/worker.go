package notifications

import (
	"log"
	"sync"
)

// Worker processes notification events in the background.
type Worker struct {
	mailer    *Mailer
	eventCh   chan NotificationEvent
	stopCh    chan struct{}
	wg        sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewWorker creates a background notification processor with a buffered channel.
func NewWorker(mailer *Mailer, bufferSize int) *Worker {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &Worker{
		mailer:  mailer,
		eventCh: make(chan NotificationEvent, bufferSize),
		stopCh:  make(chan struct{}),
	}
}

// Start begins processing events in a background goroutine.
func (w *Worker) Start() {
	select {
	case <-w.stopCh:
		return
	default:
	}

	w.startOnce.Do(func() {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.run()
		}()
	})
}

// Stop gracefully shuts down the worker, draining any remaining queued events.
func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopCh)
		w.wg.Wait()
	})
}

// Enqueue adds a notification event to the processing queue.
func (w *Worker) Enqueue(evt NotificationEvent) {
	select {
	case <-w.stopCh:
		log.Printf("[notifications] worker stopped, dropping event %s", evt.Type)
		return
	default:
	}

	select {
	case <-w.stopCh:
		log.Printf("[notifications] worker stopped, dropping event %s", evt.Type)
	case w.eventCh <- evt:
	}
}

func (w *Worker) run() {
	for {
		select {
		case <-w.stopCh:
			// Drain remaining events in the buffer for graceful shutdown
			for {
				select {
				case evt := <-w.eventCh:
					w.processEvent(evt)
				default:
					return
				}
			}
		case evt := <-w.eventCh:
			w.processEvent(evt)
		}
	}
}

func (w *Worker) processEvent(evt NotificationEvent) {
	if len(evt.Recipients) == 0 {
		return
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		log.Printf("[notifications] failed to render email for %s: %v", evt.Type, err)
		return
	}

	if w.mailer != nil {
		if err := w.mailer.Send(evt.Recipients, subject, htmlBody, textBody); err != nil {
			log.Printf("[notifications] failed to send email for %s to %v: %v", evt.Type, evt.Recipients, err)
			return
		}
	}
}
