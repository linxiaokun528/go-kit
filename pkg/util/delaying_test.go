package util_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go-kit/pkg/util"
)

type delayingHelper struct {
	ch chan int
}

func (d *delayingHelper) execute() {
	d.ch <- 0
}

var _ = Describe("DelayingExecutor", func() {
	var delayingExecutor *util.DelayingExecutor
	var helper1 delayingHelper
	var helper2 delayingHelper
	var maxDeviation time.Duration
	var delayingTime1 time.Duration
	var delayingTime2 time.Duration

	BeforeEach(func() {
		delayingExecutor = util.NewDelayingExecutor(5)
		helper1 = delayingHelper{
			ch: make(chan int, 1),
		}
		helper2 = delayingHelper{
			ch: make(chan int, 1),
		}
		maxDeviation = 100 * time.Millisecond
	})

	It("can execute a task after a specified time", func() {
		delayingExecutor.ExcuteAfter(helper1.execute, delayingTime1)
		start := time.Now()
		<-helper1.ch
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation))
	})

	It("can work with multiple tasks.", func() {
		delayingExecutor.ExcuteAfter(helper1.execute, delayingTime1)
		delayingExecutor.ExcuteAfter(helper2.execute, delayingTime2)
		start := time.Now()
		<-helper1.ch
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation))
		<-helper2.ch
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime2), maxDeviation))
	})

	It("can still work even when tasks panic.", func() {
		delayingExecutor.ExcuteAfter(func() {
			panic("test")
		}, 0)

		time.Sleep(maxDeviation)
		delayingExecutor.ExcuteAfter(func() {
			panic("test")
		}, delayingTime1)

		time.Sleep(delayingTime1 + maxDeviation)

		delayingExecutor.ExcuteAfter(helper1.execute, 0)
		time.Sleep(maxDeviation)
		Expect(helper1.ch).To(HaveLen(1))
	})

	It("can shut down immediately.", func() {
		delayingExecutor.ExcuteAfter(helper1.execute, delayingTime1)
		delayingExecutor.ShutDownFast()
		time.Sleep(maxDeviation)
		Expect(func() {
			delayingExecutor.ExcuteAfter(helper2.execute, delayingTime2)
		}).To(Panic())
		time.Sleep(delayingTime1)
		Expect(helper1.ch).To(HaveLen(0))
	})

	It("can shut down after executing remaining tasks.", func() {
		delayingExecutor.ExcuteAfter(helper1.execute, delayingTime1)
		start := time.Now()
		delayingExecutor.ShutDownWithDrain(true)
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation),
			"DelayingExecutor should blocks until all tasks are executed.")
		Expect(func() {
			delayingExecutor.ExcuteAfter(helper2.execute, delayingTime1)
		}).To(Panic())
		time.Sleep(maxDeviation) // helper1.execute is executed in a go routine, we need some time to let it finish
		Expect(helper1.ch).To(HaveLen(1))
		Expect(helper2.ch).To(HaveLen(0))
	})

	It("can shut down and executing remaining tasks in the backend.", func() {
		delayingExecutor.ExcuteAfter(helper1.execute, delayingTime1)
		start := time.Now()
		delayingExecutor.ShutDownWithDrain(false)
		Expect(time.Now()).To(BeTemporally("~", start, maxDeviation),
			"DelayingExecutor should not block.")
		time.Sleep(delayingTime1 + maxDeviation)
		Expect(helper1.ch).To(HaveLen(1))

		Expect(func() {
			delayingExecutor.ExcuteAfter(helper2.execute, delayingTime1)
		}).To(Panic())
		time.Sleep(maxDeviation) // helper1.execute is executed in a go routine, we need some time to let it finish
		Expect(helper2.ch).To(HaveLen(0))
	})

	It("won't panic when shut down after ShutDownFast.", func() {
		delayingExecutor.ShutDownFast()
		Expect(delayingExecutor.ShutDownFast).NotTo(Panic())
		Expect(func() {
			delayingExecutor.ShutDownWithDrain(false)
		}).NotTo(Panic())
		Expect(func() {
			delayingExecutor.ShutDownWithDrain(true)
		}).NotTo(Panic())
	})

	It("won't panic when shut down after ShutDownWithDrain.", func() {
		delayingExecutor.ShutDownWithDrain(false)
		Expect(delayingExecutor.ShutDownFast).NotTo(Panic())
		Expect(func() {
			delayingExecutor.ShutDownWithDrain(true)
		}).NotTo(Panic())
		Expect(func() {
			delayingExecutor.ShutDownWithDrain(false)
		}).NotTo(Panic())
	})
})

var _ = Describe("DelayingChannel", func() {
	var ch *util.DelayingChannel[int]
	var maxDeviation time.Duration
	var delayingTime1 time.Duration
	var delayingTime2 time.Duration

	BeforeEach(func() {
		ch = util.NewDelayingChannel[int](5)
		maxDeviation = 100 * time.Millisecond
		delayingTime1 = 500 * time.Millisecond
		delayingTime2 = 800 * time.Millisecond
	})

	It("can get what it adds after a specified time", func() {
		ch.AddAfter(1, delayingTime1)
		start := time.Now()
		Expect(ch.Get()).To(Equal(1))
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation))
	})

	It("can work with multiple addition.", func() {
		ch.AddAfter(1, delayingTime1)
		ch.AddAfter(2, delayingTime2)
		start := time.Now()
		Expect(ch.Get()).To(Equal(1))
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation))
		Expect(ch.Get()).To(Equal(2))
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime2), maxDeviation))
	})

	It("can get remaining items after closed.", func() {
		ch.AddAfter(1, delayingTime1)
		ch.Close()
		start := time.Now()
		Expect(ch.Get()).To(Equal(1))
		Expect(time.Now()).To(BeTemporally("~", start.Add(delayingTime1), maxDeviation))
	})

	It("can't add new items after closed.", func() {
		ch.Close()
		Expect(func() { ch.AddAfter(1, delayingTime1) }).To(Panic())
	})

	It("can't be closed multiple times.", func() {
		ch.Close()
		Expect(ch.Close).To(Panic())
	})

	It("can return a zero value after closed.", func() {
		delayingTime := 500 * time.Millisecond
		ch.AddAfter(1, delayingTime)
		ch.Close()
		Expect(ch.Get()).To(Equal(1))
		Expect(ch.Get()).To(Equal(0))

		anyCh := util.NewDelayingChannel[any](5)
		anyCh.Close()
		Expect(anyCh.Get()).To(BeNil())
		Expect(anyCh.Get()).To(BeNil())
	})
})
