package goroutine_pool

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"

	"golang.org/x/sync/errgroup"
)

// GoroutinePool 结构体包含一个等待组（sync.WaitGroup）和一个通道（chan struct{}）来控制并发数
type GoroutinePool struct {
	wg sync.WaitGroup
	Ch chan struct{}
}

// NewGoroutinePool 创建一个新的GoroutinePool实例，最大并发数由maxGoroutines参数指定
func NewGoroutinePool(maxGoroutines int) *GoroutinePool {
	return &GoroutinePool{
		Ch: make(chan struct{}, maxGoroutines),
	}
}

func ParallN(ctx context.Context, n int, tasks <-chan func(ctx context.Context) error) error {
	var eg errgroup.Group
	eg.SetLimit(n)

	for i := 0; i < n; i++ {
		eg.Go(func() error {
			for t := range tasks {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					err := t(ctx)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
	}

	return eg.Wait()
}

func ParallN2(ctx context.Context, max int, task []func(ctx context.Context) error) error {
	eg, ctx := errgroup.WithContext(ctx)
	tasksChannels := make(chan func(ctx context.Context) error, max*10)
	go func() {
		defer close(tasksChannels)
		for _, t := range task {
			select {
			case <-ctx.Done():
				return
			case tasksChannels <- t:
			}
		}
	}()

	for i := 0; i < max; i++ {
		eg.Go(func() error {
			for t := range tasksChannels {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					err := t(ctx)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
	}

	return eg.Wait()
}

// Execute 启动一个新的goroutine，如果通道满了，则等待
func (p *GoroutinePool) Execute(f func(param interface{}), param interface{}) {
	p.wg.Add(1)
	p.Ch <- struct{}{} // 占用一个槽位
	go func() {
		// defer 异常处理
		defer func() {
			if e := recover(); e != nil {
				log.Context(context.Background()).Errorf("goroutine panic err:%v", e)
				return
			}
		}()
		defer p.wg.Done()
		defer func() { <-p.Ch }()
		f(param) // 执行传入的函数
	}()
}

// Wait 等待所有goroutine执行完毕
func (p *GoroutinePool) Wait() {
	p.wg.Wait()
}
