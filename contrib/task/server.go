package task

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"fission-basic/contrib/internel"
	taskq "fission-basic/kit/task"
)

type Task struct {
	Queue  *taskq.Queue
	Func   func(ctx context.Context, ids []string) error
	Number uint // 每次从队列拉取最大数量
	D      time.Duration
}

func (t *Task) Clone() *Task {
	return &Task{
		Queue:  t.Queue,
		Func:   t.Func,
		Number: t.Number,
		D:      t.D,
	}
}

func (t *Task) runOnce(ctx context.Context) error {
	number := t.Number
	if number == 0 {
		number = 1
	}

	ids, err := t.Queue.Receive(number, t.D)
	if err != nil {
		return err
	}
	// del set
	defer t.Queue.Release(ids)
	lenth := len(ids)

	tracer := tracing.NewTracer(trace.SpanKindServer)
	ctx, span := tracer.Start(ctx, "consumer", internel.NewTextMap())
	defer func() {
		tracer.End(ctx, span, "", nil)
	}()

	err = t.Func(ctx, ids)
	if err != nil {
		return err
	}

	if lenth == 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return nil
}

type Server struct {
	cancel context.CancelFunc
	sig    chan struct{}
	tasks  []*Task
}

func NewServer() *Server {
	s := Server{}
	return &s
}

func (s *Server) AddTask(t *Task) error {
	s.tasks = append(s.tasks, t)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	var eg errgroup.Group
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	for i := range s.tasks {
		t := s.tasks[i]
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Worker received cancel signal, exiting...")
					return ctx.Err()
				default:
					// 继续执行
				}

				err := t.runOnce(ctx)
				if err != nil {
					return err
				}
			}
		})
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	return nil
}
