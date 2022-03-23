package main

import (
	"log"
	"strings"
	"time"
	"github.com/getsentry/sentry-go"
)

func RecoverRepanic(f func()) {
	// Clone the current hub so that modifications of the scope are visible only
	// within this function.
	hub := sentry.CurrentHub().Clone()

	// filterFrames removes frames from outgoing events that reference the
	// RecoverRepanic function and its subfunctions.
	filterFrames := func(event *sentry.Event) {
		for _, e := range event.Exception {
			if e.Stacktrace == nil {
				continue
			}
			frames := e.Stacktrace.Frames[:0]
			for _, frame := range e.Stacktrace.Frames {
				if frame.Module == "main" && strings.HasPrefix(frame.Function, "RecoverRepanic") {
					continue
				}
				frames = append(frames, frame)
			}
			e.Stacktrace.Frames = frames
		}
	}

	// Add an EventProcessor to the scope. The event processor is a function
	// that can change events before they are sent to Sentry.
	// Alternatively, see also ClientOptions.BeforeSend, which is a special
	// event processor applied to all events.
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			filterFrames(event)
			return event
		})
	})

	// See https://golang.org/ref/spec#Handling_panics.
	// This will recover from runtime panics and then panic again after
	// reporting to Sentry.
	defer func() {
		if x := recover(); x != nil {
			log.Printf("RECOVER %+v",x )
			// Create an event and enqueue it for reporting.
			hub.Recover(x)
			// Because the goroutine running this code is going to crash the
			// program, call Flush to send the event to Sentry before it is too
			// late. Set the timeout to an appropriate value depending on your
			// program. The value is the maximum time to wait before giving up
			// and dropping the event.
			hub.Flush(2 * time.Second)
			// Note that if multiple goroutines panic, possibly only the first
			// one to call Flush will succeed in sending the event. If you want
			// to capture multiple panics and still crash the program
			// afterwards, you need to coordinate error reporting and
			// termination differently.
			panic(x)
		}
	}()

	// Run the original function.
	f()
}
