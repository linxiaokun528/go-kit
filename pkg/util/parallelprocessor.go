package util

import (
	"context"
	"fmt"
	"sync"
)

type LoopFunc func(ctx context.Context) bool
type PanicHandler func(r any)

type ParallelProcessor struct {
	loopFunc     LoopFunc
	panicHandler PanicHandler
	// If we don't mind relying on k8s library, we can use k8s.io/apimachinery/pkg/util.Group
	wait sync.WaitGroup
}

func NewParallelProcessor(loopFunc LoopFunc, panicHandler PanicHandler) *ParallelProcessor {
	return &ParallelProcessor{
		loopFunc:     loopFunc,
		panicHandler: panicHandler,
		wait:         sync.WaitGroup{},
	}
}

// Start : blocks until ctx is done or loopFunc returns false in all routines
func (p *ParallelProcessor) Start(consumerNum int, ctx context.Context) {
	if consumerNum <= 0 {
		panic(fmt.Errorf("consumerNum should be positive"))
	}

	p.wait.Add(consumerNum)
	for i := 0; i < consumerNum; i++ {
		go func() {
			defer p.wait.Done()
			for p.worker(ctx) {

			}
		}()
	}
	p.wait.Wait()
}

func (p *ParallelProcessor) worker(ctx context.Context) (goNext bool) {
	defer func() {
		if r := recover(); r != nil { // in case a panic happens while handling panics
			goNext = true
		}
	}()

	if p.panicHandler != nil {
		defer func() {
			if r := recover(); r != nil {
				p.panicHandler(r)
			}
		}()
	}

	select {
	case <-ctx.Done():
		return false
	default:
		return p.loopFunc(ctx)
	}
}

type ProducerFunc[T any] func(ctx context.Context) T
type ConsumerFunc[T any] func(product T, ctx context.Context)
type ParallelConsumingProcessor[T any] struct {
	producerFunc ProducerFunc[T]
	consumerFunc ConsumerFunc[T]
	processor    *ParallelProcessor
}

func NewParallelConsumingProcessor[T any](producerFunc ProducerFunc[T], consumerFunc ConsumerFunc[T],
	panicHandler PanicHandler) *ParallelConsumingProcessor[T] {
	result := ParallelConsumingProcessor[T]{
		producerFunc: producerFunc,
		consumerFunc: consumerFunc,
	}
	result.processor = NewParallelProcessor(result.process, panicHandler)
	return &result
}

func (p *ParallelConsumingProcessor[T]) Start(consumerNum int, ctx context.Context) {
	p.processor.Start(consumerNum, ctx)
}

func (p *ParallelConsumingProcessor[T]) process(ctx context.Context) bool {
	// Maybe use a channel like the following, so that producer doesn't need to be thread-safe
	// channel := make(chan T)
	// go func() {
	// 	for true {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			channel <- p.produce()
	// 		}
	// 	}
	// }()
	//
	// go func() {
	// 	for true {
	// 		select {
	// 		case <-ctx.Done():
	// 			return;
	// 		default:
	// 			p.consume(<-channel)
	// 		}
	// 	}
	// }()

	var product T

	select {
	case <-ctx.Done():
		return false
	default:
		product = p.producerFunc(ctx)
	}

	select {
	case <-ctx.Done():
		return false
	default:
		p.consumerFunc(product, ctx)
	}

	return true
}
