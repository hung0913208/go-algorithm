package utils

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	sentry "github.com/getsentry/sentry-go"
	trace "github.com/uptrace/uptrace-go/uptrace"

	"github.com/hung0913208/go-algorithm/lib/container"
	"github.com/hung0913208/go-algorithm/lib/logs"
)

type Handler func(w http.ResponseWriter, r *http.Request)

func DoApi(module, method, api string) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			return
		}

		logger := logs.NewLoggerWithStacktrace()
		ctx := r.Context()

		tracer := container.NewRootTrace(api, ctx)
		span, _ := tracer.Enter(api)
		defer tracer.Exit()

		defer func() {
			err := recover()

			if err != nil {
				span.RecordError(fmt.Errorf("Crash %v", err))

				if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
					logger.Warnf(
						"Crash %v",
						err,
					)
				}

				container.Crash(w, fmt.Errorf("Crash %v", err))

				debug.PrintStack()
			}

			sentry.Flush(2 * time.Second)
			trace.ForceFlush(ctx)
		}()

		container.HandleRESTfulAPIs(module, w, r)
	}
}
