package util_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go-kit/pkg/util"
	"go-kit/pkg/util/collection"
)

type loopFuncHelper struct {
	invokedTime int
	locker      *sync.Mutex
}

func newLoopFuncHelper() *loopFuncHelper {
	return &loopFuncHelper{
		invokedTime: 0,
		locker:      &sync.Mutex{},
	}
}

func (l *loopFuncHelper) invokeOnce(ctx context.Context) bool {
	defer l.locker.Unlock()
	l.locker.Lock()

	l.invokedTime += 1
	return false
}

func (l *loopFuncHelper) invokeInfinitely(ctx context.Context) bool {
	defer l.locker.Unlock()
	l.locker.Lock()

	l.invokedTime += 1
	return true
}

var doNothingHandler = func(r any) {}

var _ = Describe("ParallelProcessor", func() {
	var processor *util.ParallelProcessor
	var helper *loopFuncHelper
	var ctx context.Context
	var cancelFunc context.CancelFunc
	// Use a chan instead of a bool variable to avoid synchronization problems
	var stopChan chan bool

	BeforeEach(func() {
		helper = newLoopFuncHelper()
		ctx, cancelFunc = context.WithCancel(context.Background())
		DeferCleanup(cancelFunc)
		stopChan = make(chan bool)
	})

	run_processor := func() {
		processor.Start(10, ctx)
		close(stopChan)
	}

	Describe("can run loopFunc in different routines.", func() {
		It("The loopFunc is invoked concurrently.", func() {
			// Don't use a bool variable here which may have synchronization problems theoretically
			var invoked uint32 = 0
			locker := sync.Mutex{}
			processor = util.NewParallelProcessor(func(ctx context.Context) bool {
				defer locker.Unlock()
				for !locker.TryLock() {
					atomic.AddUint32(&invoked, 1)
				}
				return helper.invokeOnce(ctx)
			}, doNothingHandler)

			Eventually(func() uint32 {
				processor.Start(10, ctx)
				return atomic.LoadUint32(&invoked)
			}).Should(BeNumerically(">", 0))
		})

		Context("The concurrency is specified by consumerNum.", func() {
			BeforeEach(func() {
				processor = util.NewParallelProcessor(helper.invokeOnce, doNothingHandler)
			})

			It("If the consumerNum is 1, then the loopFunc is only invoked once", func() {
				processor.Start(1, ctx)
				Expect(helper.invokedTime).To(Equal(1))
			})

			It("If the consumerNum is 3, then the loopFunc is invoked three times", func() {
				processor.Start(3, ctx)
				Expect(helper.invokedTime).To(Equal(3))
			})

			It("If the consumerNum is 0, a panic will happen", func() {
				Expect(func() { processor.Start(0, ctx) }).To(Panic())
			})

			It("If the consumerNum is negative, a panic will happen", func() {
				Expect(func() { processor.Start(-1, ctx) }).To(Panic())
			})
		})
	})

	Describe("stops if the context is done.", func() {
		BeforeEach(func() {
			processor = util.NewParallelProcessor(helper.invokeInfinitely, doNothingHandler)
		})

		It("If the ctx is done before start, the loopFunc won't be invoked.", func() {
			cancelFunc()
			processor.Start(10, ctx)
			Expect(helper.invokedTime).To(Equal(0))
		})

		It("If the ctx is done after start, the loopFunc will be invoked, but stopped by the cxt.", func() {
			go run_processor()
			// make sure the loopFunc is invoked
			Eventually(func() int { return helper.invokedTime }).ShouldNot(BeZero())
			cancelFunc()
			Eventually(func() <-chan bool { return stopChan }).Should(BeClosed())
			invokedTime := helper.invokedTime
			Consistently(func() int { return helper.invokedTime }).Should(Equal(invokedTime))
		})
	})

	Describe("can handle panics with the panicHandler.", func() {
		var actualErr error
		var expectedErr error
		BeforeEach(func() {
			expectedErr = fmt.Errorf("panic for test")
			actualErr = nil
			processor = util.NewParallelProcessor(
				func(ctx context.Context) bool {
					helper.invokeOnce(ctx)
					panic(expectedErr)
				},
				func(r any) {
					actualErr = r.(error)
				})
		})

		It("panicHandler can get the value thrown by panic.", func() {
			go run_processor()
			Eventually(func() error { return actualErr }).Should(MatchError(expectedErr))
		})
	})

	Describe("won't stop running even if panicHandler panics.", func() {
		BeforeEach(func() {
			processor = util.NewParallelProcessor(
				func(ctx context.Context) bool {
					helper.invokeOnce(ctx)
					panic(fmt.Errorf("panic for test"))
				},
				func(r any) {
					panic(r)
				})
		})

		It("panic value thrown by panicHandler will be ignored.", func() {
			go run_processor()
			// make sure the loopFunc is invoked
			Eventually(func() int { return helper.invokedTime }).ShouldNot(BeZero())
			Eventually(func() chan bool { return stopChan }).ShouldNot(BeClosed())
		})
	})
})

type producer struct {
	invokedTimes        int
	isInfinite          bool
	expectedInvokedTime int
	cancelFunc          context.CancelFunc
	locker              sync.Locker
}

func (p *producer) produce(ctx context.Context) int {
	defer p.locker.Unlock()
	p.locker.Lock()
	p.invokedTimes += 1
	if !p.isInfinite && p.invokedTimes >= p.expectedInvokedTime {
		p.cancelFunc()
	}
	return p.invokedTimes
}

func (p *producer) GetInvokedTimes() int {
	defer p.locker.Unlock()
	p.locker.Lock()

	return p.invokedTimes
}

func newProducer(expectedInvokedTime int, cancelFunc context.CancelFunc) *producer {
	return &producer{
		invokedTimes:        0,
		isInfinite:          false,
		expectedInvokedTime: expectedInvokedTime,
		cancelFunc:          cancelFunc,
		locker:              &sync.Mutex{}}
}

func newInfiniteProducer() *producer {
	return &producer{
		invokedTimes:        0,
		isInfinite:          true,
		expectedInvokedTime: 0,
		cancelFunc:          nil,
		locker:              &sync.Mutex{}}
}

type consumer struct {
	results collection.Set[int]
}

func (c *consumer) consume(product int, ctx context.Context) {
	c.results.Add(product)
}

func (c *consumer) getResults() []int {
	results := c.results.ToArray()
	sort.Ints(results)
	return results
}

func newConsumer() *consumer {
	return &consumer{results: collection.NewThreadSafeSet[int, int](func(value int) int { return value },
		func(t1, t2 int) bool { return t1 == t2 })}
}

var _ = Describe("ParallelConsumingProcessor", func() {
	var producer *producer
	var consumer *consumer
	var producerFunc util.ProducerFunc[int]
	var consumerFunc util.ConsumerFunc[int]
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var processor *util.ParallelConsumingProcessor[int]

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.Background())
		producer = newInfiniteProducer()
		consumer = newConsumer()
		producerFunc = producer.produce
		consumerFunc = consumer.consume
	})

	Context("let consumers consume things produced by producers.", func() {
		BeforeEach(func() {
			producer = newProducer(10, cancelFunc)
			producerFunc = producer.produce
			processor = util.NewParallelConsumingProcessor[int](producerFunc, consumerFunc, doNothingHandler)
		})

		It("It works fine when there is only one consumer.", func() {
			processor.Start(1, ctx)
			Expect(consumer.getResults()).To(Equal([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}))
		})

		It("It works fine when there are more than one consumers.", func() {
			processor.Start(2, ctx)
			Expect(consumer.getResults()).To(Equal([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}))
		})
	})

	It("let consumers work concurrently.", func() {
		// Don't use a bool variable here which may have synchronization problems theoretically
		var invoked uint32 = 0
		locker := sync.Mutex{}
		processor := util.NewParallelConsumingProcessor[int](producerFunc, func(product int, ctx context.Context) {
			defer locker.Unlock()
			for !locker.TryLock() {
				atomic.AddUint32(&invoked, 1)
			}
			consumerFunc(product, ctx)
		}, doNothingHandler)

		// Don't simply use`go processor.Start(10, ctx)`, because a backend routine may still running after
		// this function ends and thus affect other test cases.
		stopCh := make(chan bool)
		go func() {
			processor.Start(10, ctx)
			close(stopCh)
		}()

		func() {
			defer cancelFunc()
			Eventually(func() uint32 { return atomic.LoadUint32(&invoked) }).Should(BeNumerically(">", 0))
		}()
		<-stopCh
	})

	Describe("stops working when ctx is done.", func() {
		It("If ctx is done before start, then producers and consumers don't even start to work.", func() {
			processor := util.NewParallelConsumingProcessor[int](producerFunc, consumerFunc, doNothingHandler)
			cancelFunc()
			processor.Start(10, ctx)
			Expect(producer.GetInvokedTimes()).To(BeZero())
			Expect(consumer.getResults()).To(BeEmpty())
		})

		It("If producers and consumers is already working, done ctx can stop the processor.", func() {
			processor := util.NewParallelConsumingProcessor[int](producerFunc, consumerFunc, doNothingHandler)

			stopCh := make(chan bool)
			go func() {
				processor.Start(10, ctx)
				close(stopCh)
			}()

			Eventually(consumer.results.Len).ShouldNot(Equal(0))
			cancelFunc()
			Eventually(stopCh).Should(BeClosed())
			values := consumer.getResults()
			Consistently(consumer.getResults).Should(Equal(values))
		})
	})

	Describe("let panicHandler handle panics", func() {
		var expectedErr any
		var actualErr any
		var panicHandler util.PanicHandler

		BeforeEach(func() {
			expectedErr = fmt.Errorf("test")
			panicHandler = func(r any) {
				actualErr = r
			}
		})

		It("of consumers", func() {
			processor := util.NewParallelConsumingProcessor(func(ctx context.Context) int {
				cancelFunc()
				panic(expectedErr)
			}, consumerFunc, panicHandler)

			processor.Start(1, ctx)
			Expect(actualErr).To(Equal(expectedErr))
		})

		It("of producers", func() {
			processor := util.NewParallelConsumingProcessor(producerFunc, func(product int, ctx context.Context) {
				cancelFunc()
				panic(expectedErr)
			}, panicHandler)

			processor.Start(1, ctx)
			Expect(actualErr).To(Equal(expectedErr))
		})
	})
})
