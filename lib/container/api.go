package container

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hung0913208/go-algorithm/lib/algorithm/heap"
	"go.opentelemetry.io/otel/trace"
)

type Module interface {
	Init(timeout time.Duration) error
	Deinit() error

	GetTimeout() time.Duration
}

type CliModule interface {
	Module

	Execute(args []string) error
}

type RestModule interface {
	Module

	SetResponseWriter(writer http.ResponseWriter)
	SetRequestReader(reader *http.Request)

	// @NOTE: define methods
	DoGet(
		function string,
		kwargs url.Values,
	) (interface{}, error)
	DoPost(
		function string,
		kwargs url.Values,
	) (interface{}, error)
}

type RpcModule interface {
	Module

	PairWith(module string) error
}

type CronModule interface {
	Module

	GetCron() map[string]int
	GetJob(name string) (func(cancelChan chan bool, wg *sync.WaitGroup), error)
}

type wrapImpl struct {
	name   string
	module Module
	index  int
	status bool
}

type containerImpl struct {
	mapping  map[string]wrapImpl
	modules  []Module
	tracer   trace.Tracer
	jobHeap  heap.Heap
	timer    int64
	sigChan  chan os.Signal
	stopChan chan bool
	doneChan chan bool
}

type jobImpl struct {
	timer    int
	handler  func(cancelChan chan bool, wg *sync.WaitGroup)
	interval int
}

type responseImpl struct {
	Description string      `json:"description"`
	Code        int         `json:"code"`
	Data        interface{} `json:"data"`
}

var iContainerManager *containerImpl

func Init(tracer trace.Tracer) error {
	if iContainerManager != nil {
		return errors.New("Only call init container one time at the begining")
	}

	sigChan := make(chan os.Signal, 1)
	stopChan := make(chan bool, 1)
	doneChan := make(chan bool, 1)
	signal.Notify(sigChan, syscall.SIGALRM)

	go func() {
		wg := &sync.WaitGroup{}
		cancelChan := make(chan bool)

		for {
			select {
			case <-sigChan:
				var job *jobImpl

				alarm(1)

				if obj, err := iContainerManager.jobHeap.Get(); err == nil {
					timer := atomic.AddInt64(&iContainerManager.timer, 1)

					if obj.(jobImpl).timer == int(timer) {
						job = obj.(*jobImpl)
						iContainerManager.jobHeap.Pop()
						iContainerManager.jobHeap.Push(&jobImpl{
							timer:    job.timer + job.interval,
							handler:  job.handler,
							interval: job.interval,
						})
					}
				}

				if job != nil {
					wg.Add(1)
					job.handler(cancelChan, wg)
				}

			case <-stopChan:
				cancelChan <- true
				wg.Wait()
				doneChan <- true
			}
		}
	}()

	alarm(1)

	iContainerManager = &containerImpl{
		mapping: make(map[string]wrapImpl),
		modules: make([]Module, 0),
		tracer:  tracer,
		jobHeap: heap.NewHeapWithComparator(func(l, r interface{}) int {
			return l.(jobImpl).timer - r.(jobImpl).timer
		}),
		sigChan:  sigChan,
		stopChan: stopChan,
		doneChan: doneChan,
	}
	return nil
}

func RegisterSimpleModule(
	name string,
	module Module,
	timeout int,
) error {
	if iContainerManager == nil {
		if err := Init(nil); err != nil {
			return err
		}
	}

	if iContainerManager == nil {
		return errors.New("Con't setup container manager")
	}

	if _, ok := iContainerManager.mapping[name]; ok {
		return fmt.Errorf("Object %s has been registered", name)
	}

	if err := module.Init(time.Duration(timeout) * time.Millisecond); err != nil {
		return err
	}

	iContainerManager.mapping[name] = wrapImpl{
		name:   name,
		module: module,
		index:  len(iContainerManager.modules),
		status: true,
	}
	iContainerManager.modules = append(iContainerManager.modules, module)
	return nil
}

func RegisterRpcModule(
	name string,
	module Module,
	timeout int,
) error {
	err := RegisterSimpleModule(name, module, timeout)
	if err != nil {
		return err
	}

	return nil
}

func RegisterCronModule(
	name string,
	module Module,
	timeout int,
) error {
	err := RegisterSimpleModule(name, module, timeout)
	if err != nil {
		return err
	}

	for name, interval := range module.(CronModule).GetCron() {
		handler, err := module.(CronModule).GetJob(name)
		if err != nil {
			return err
		}

		if handler != nil {
			iContainerManager.jobHeap.Push(&jobImpl{
				timer:    interval,
				handler:  handler,
				interval: interval,
			})
		}
	}
	return nil
}

func Terminate(msg string, exitCode int) {
	fmt.Printf("Exit(%d) with error %s", exitCode, msg)

	if iContainerManager != nil {
		for _, wrap := range iContainerManager.mapping {
			if !wrap.status {
				continue
			}

			if err := wrap.module.Deinit(); err != nil {
				log.Fatalf("%v", err)
			}

			wrap.status = false
		}
	}

	if exitCode != 0 {
		panic(fmt.Sprintf("Exit(%d) with error %s", exitCode, msg))
	}
}

func Pick(name string) (Module, error) {
	wrapper, ok := iContainerManager.mapping[name]

	if !ok {
		return nil, fmt.Errorf("Module `%s` doesn`t exist", name)
	}
	return wrapper.module, nil
}

func Lookup(index int) (Module, error) {
	if index >= len(iContainerManager.modules) {
		return nil, fmt.Errorf(
			"index `%d` is out of scope, must below %d",
			index,
			len(iContainerManager.modules),
		)
	}

	return iContainerManager.modules[index], nil
}

func Crash(writer http.ResponseWriter, reason error) {
	resp := &responseImpl{}

	writer.WriteHeader(http.StatusInternalServerError)
	resp.Description = fmt.Sprintf("Crash %v", reason)
	resp.Code = 502

	out, _ := json.Marshal(resp)
	writer.Write(out)
}

func HandleRESTfulAPIs(
	module string,
	writer http.ResponseWriter, reader *http.Request,
) error {
	tracer := NewTrace(reader.URL.Path)
	span, _ := tracer.Enter("HandleRESfulAPIs")
	resp := &responseImpl{}

	defer tracer.Exit()

	wrapper, err := Pick(module)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		resp.Description = fmt.Sprintf("%v", err)
		resp.Code = 500

		if out, err := json.Marshal(resp); err != nil {
			span.RecordError(err)
			return err
		} else {
			writer.Write(out)
		}

		span.RecordError(err)
		return err
	}

	if restObj, ok := wrapper.(RestModule); !ok {
		err = fmt.Errorf("module `%s` isn't RESTful module", module)
		writer.WriteHeader(http.StatusInternalServerError)
		resp.Description = fmt.Sprintf("%v", err)
		resp.Code = 500

		span.RecordError(err)

		if out, err := json.Marshal(resp); err != nil {
			span.RecordError(err)
			return err
		} else {
			writer.Write(out)
		}

		return fmt.Errorf("module `%s` isn't RESTful module", module)
	} else {
		var finalErr error

		restObj.SetResponseWriter(writer)
		restObj.SetRequestReader(reader)

		switch reader.Method {
		case "GET":
			body, err := restObj.DoGet(reader.URL.Path, reader.URL.Query())
			resp := &responseImpl{}

			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				resp.Description = fmt.Sprintf("%v", err)
				resp.Code = 500
				finalErr = err

				span.RecordError(err)
			} else {
				resp.Data = body
				resp.Code = 200
			}

			if out, err := json.Marshal(resp); err != nil {
				span.RecordError(err)
				return err
			} else {
				writer.Write(out)
			}

			finalErr = err

		case "POST":
			body, err := restObj.DoPost(reader.URL.Path, reader.URL.Query())
			resp := &responseImpl{}

			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				resp.Description = fmt.Sprintf("%v", err)
				resp.Code = 500

				span.RecordError(err)
			} else {
				resp.Data = body
				resp.Code = 200
			}

			if out, err := json.Marshal(resp); err != nil {
				span.RecordError(err)
				return err
			} else {
				writer.Write(out)
			}

			finalErr = err

		default:
			resp := &responseImpl{}

			writer.WriteHeader(http.StatusInternalServerError)
			resp.Description = "not support this method"
			resp.Code = 404

			out, err := json.Marshal(resp)
			if err != nil {
				span.RecordError(err)
				return err
			} else {
				writer.Write(out)
			}

			finalErr = errors.New("Not support this method")
		}
		return finalErr
	}
}

func HandleCliExec(
	module string, args []string,
) error {
	trace := NewTrace(module)
	span, spanCtx := trace.Enter("HandleCliExec")

	defer trace.Exit()

	wrapper, err := Pick(module)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if cliObj, ok := wrapper.(CliModule); !ok {
		return fmt.Errorf("module `%s` is not cli module", module)
	} else {
		timeoutCtx, cancel := context.WithTimeout(
			context.Background(),
			cliObj.GetTimeout()*time.Millisecond,
		)
		errCh := make(chan error, 1)

		defer cancel()

		go func(ctx context.Context) {
			tracer := NewRootTrace("gorouting", ctx)
			tracer.Enter("gorouting")

			defer tracer.Exit()
			errCh <- cliObj.Execute(args)
		}(spanCtx)

		select {
		case <-timeoutCtx.Done():
			err := fmt.Errorf("Timeout perform `%s`, args = %s", module, args)
			span.RecordError(err)
			return err
		case result := <-errCh:
			if result != nil {
				span.RecordError(result)
			}
			return result
		}
	}
}
