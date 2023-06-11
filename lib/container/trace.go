package container

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/golang-collections/collections/stack"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Trace interface {
	Enter(fn string) (trace.Span, context.Context)
	Exit()

	GetName() string
	GetTrace() trace.Tracer
	GetContext() context.Context
}

type traceImpl struct {
	context context.Context
	tracer  trace.Tracer
	flow    *stack.Stack
	name    string
	enabled bool
}

type trackImpl struct {
	row, col int
	line     string
	root     Trace
}

var iTraceManager map[string]Trace
var iTrackPoints []trackImpl

type spanImpl struct {
	tracer   *traceImpl
	parent   context.Context
	selflink context.Context
	span     trace.Span
}

var (
	_ trace.Span = &spanImpl{}
)

func NewRootTrace(
	name string,
	context context.Context,
	tracer ...trace.Tracer,
) Trace {
	if iTraceManager == nil {
		iTraceManager = make(map[string]Trace)
	}

	if iTrackPoints == nil {
		iTrackPoints = make([]trackImpl, 0)
	}

	if _, ok := iTraceManager[name]; !ok {
		lines := strings.Split(string(debug.Stack()), "\n")

		if len(tracer) > 0 {
			iTraceManager[name] = &traceImpl{
				context: context,
				tracer:  tracer[0],
				flow:    stack.New(),
				name:    name,
				enabled: true,
			}
		} else {
			iTraceManager[name] = &traceImpl{
				context: context,
				tracer:  iContainerManager.tracer,
				flow:    stack.New(),
				name:    name,
				enabled: true,
			}
		}

		for i, str := 1, lines[5]; i <= len(str); i++ {
			if str[len(lines[5])-i] == '(' {
				iTrackPoints = append(iTrackPoints, trackImpl{
					row:  len(lines) - 5,
					col:  len(str) - i,
					root: iTraceManager[name],
					line: str[0 : len(str)-i],
				})
				break
			}
		}
	}

	return iTraceManager[name]
}

func NewConcurrentTrace(index int, parent Trace) Trace {
	name := fmt.Sprintf("concurrent:%s:%d", parent.GetName(), index)

	if iTraceManager == nil {
		iTraceManager = make(map[string]Trace)
	}

	if iTrackPoints == nil {
		iTrackPoints = make([]trackImpl, 0)
	}

	if _, ok := iTraceManager[name]; !ok {
		lines := strings.Split(string(debug.Stack()), "\n")

		iTraceManager[name] = &traceImpl{
			tracer:  parent.GetTrace(),
			context: parent.GetContext(),
			flow:    stack.New(),
			name:    name,
			enabled: true,
		}

		for i, str := 1, lines[5]; i <= len(str); i++ {
			if str[len(lines[5])-i] == '(' {
				iTrackPoints = append(iTrackPoints, trackImpl{
					row:  len(lines) - 5,
					col:  len(str) - i,
					root: iTraceManager[name],
					line: str[0 : len(str)-i],
				})
				break
			}
		}
	}

	return iTraceManager[name]
}

func NewTrace(name string) Trace {
	lines := strings.Split(string(debug.Stack()), "\n")

	// @TODO: use aho-corasick to resolve this one with liner bigO
	for _, track := range iTrackPoints {
		if len(lines) <= track.row {
			continue
		}

		str := lines[len(lines)-track.row]

		if len(str) <= track.col {
			continue
		}

		if str[track.col] != '(' {
			continue
		}

		if str[0:track.col] != track.line {
			continue
		}

		return track.root
	}

	return &traceImpl{
		enabled: false,
	}
}

func GetContext() context.Context {
	lines := strings.Split(string(debug.Stack()), "\n")

	// @TODO: use aho-corasick to resolve this one with liner bigO
	for _, track := range iTrackPoints {
		if len(lines) <= track.row {
			continue
		}

		str := lines[len(lines)-track.row]

		if len(str) <= track.col {
			continue
		}

		if str[track.col] != '(' {
			continue
		}

		if str[0:track.col] != track.line {
			continue
		}

		return track.root.(*traceImpl).GetContext()
	}

	return context.Background()
}

func (self *traceImpl) Enter(function string) (trace.Span, context.Context) {
	if !self.enabled || len(os.Getenv("UPTRACE_DSN")) == 0 || self.tracer == nil {
		return &spanImpl{}, nil
	}

	parent := self.GetContext()
	current, span := self.tracer.Start(parent, function)

	self.flow.Push(&spanImpl{
		span:     span,
		tracer:   self,
		parent:   parent,
		selflink: current,
	})

	return span, current
}

func (self *traceImpl) Exit() {
	if !self.enabled || len(os.Getenv("UPTRACE_DSN")) == 0 || self.tracer == nil {
		return
	}

	if self.flow.Len() > 0 {
		obj := self.flow.Peek()
		if spanObj, ok := obj.(*spanImpl); ok {
			spanObj.span.End()
		}

		self.flow.Pop()
	}
}

func (self *traceImpl) GetName() string {
	return self.name
}

func (self *traceImpl) GetTrace() trace.Tracer {
	return self.tracer
}

func (self *traceImpl) GetContext() context.Context {
	if self.flow.Len() > 0 {
		obj := self.flow.Peek()
		if spanObj, ok := obj.(*spanImpl); ok {
			return spanObj.selflink
		}
	}

	return self.context
}

func (self *traceImpl) getSpan() trace.Span {
	if self.flow.Len() > 0 {
		obj := self.flow.Peek()
		if spanObj, ok := obj.(*spanImpl); ok {
			return spanObj.span
		}
	}

	return nil
}

func (self *spanImpl) End(options ...trace.SpanEndOption) {}

func (self *spanImpl) AddEvent(name string, options ...trace.EventOption) {}

func (self *spanImpl) IsRecording() bool { return false }

func (self *spanImpl) RecordError(err error, options ...trace.EventOption) {}

func (self *spanImpl) SpanContext() trace.SpanContext { return trace.SpanContext{} }

func (self *spanImpl) SetStatus(code codes.Code, description string) {}

func (self *spanImpl) SetName(name string) {}

func (self *spanImpl) SetAttributes(kv ...attribute.KeyValue) {}

func (self *spanImpl) TracerProvider() trace.TracerProvider { return nil }
