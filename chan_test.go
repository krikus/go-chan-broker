package gochanbroker

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var noop = func(_ string) bool { return true }

var _ = Describe("ChanBroker basic cases", func() {
	DescribeTable("Check constructor", func(c int) {
		chb := CreateChanBroker(c, noop)
		Expect(chb.concurrency).To(Equal(c))
	},
		Entry("Two channels", 2),
		Entry("Five channels", 5))

	When("Empty ChanBroker is closed", func() {
		chb := CreateChanBroker(1, noop)

		It("Result chan should be closed", func() {
			closedByBroker := make(chan bool, 1)

			go func() {
				select {
				case res := <-chb.GetResultsChan():
					Expect(res).To(BeNil())
					closedByBroker <- true
				case <-time.After(1 * time.Second):
					closedByBroker <- false
				}
			}()

			chb.Finalize()
			ret := <-closedByBroker
			Expect(ret).To(Equal(true))
		})
	})
})

var _ = Describe("Processing goroutines", func() {
	DescribeTable("Processing with different thread count", func(concurrency int, numberOfJobs int) {
		defer GinkgoRecover()

		timePerJob := 10 * time.Millisecond
		chb := CreateChanBroker(concurrency, func(s string) bool {
			<-time.After(timePerJob)
			return strings.Index(s, "Job ") == 0
		})

		counter := make(chan int, 1)

		go func() {
			num := 0
			for res := range chb.GetResultsChan() {
				num++
				Expect(res.Result).To(Equal(true))
			}
			counter <- num
			close(counter)
		}()

		startTime := time.Now()
		for i := 0; i < numberOfJobs; i++ {
			chb.AddJob(fmt.Sprintf("Job %d", i))
		}

		chb.Finalize()
		ret := <-counter
		testDuration := time.Since(startTime)

		Expect(ret).To(Equal(numberOfJobs))
		Expect(testDuration.Milliseconds()).To(BeNumerically("~", timePerJob.Milliseconds()*int64(numberOfJobs)/int64(concurrency), timePerJob.Milliseconds()*2))

	},
		Entry("One thread", 1, 10),
		Entry("Two threads", 2, 20),
		Entry("Ten threads", 10, 50),
	)
})

var _ = Describe("GetResultsChan()", func() {
	When("Processing jobs", func() {
		It("Should return results in proper order", func() {
			jobs := []time.Duration{100 * time.Millisecond, 1 * time.Millisecond, 50 * time.Millisecond}
			chb := CreateChanBroker(99, func(t time.Duration) bool {
				<-time.After(t)
				return t.Milliseconds()%2 == 0
			})

			go func() {
				for _, job := range jobs {
					chb.AddJob(job)
				}

				chb.Finalize()
			}()

			index := 0
			for result := range chb.GetResultsChan() {
				Expect(result.Result).To(Equal(result.Key.Milliseconds()%2 == 0))
				Expect(result.Key).To(Equal(jobs[index]))
				index++
			}
		})
	})
})
