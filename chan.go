package gochanbroker

import (
	"sync"
)

type ChanResult[K comparable, R any] struct {
	Key    K
	Result R
	final  bool
}

// ChanBroker - broker instance
type ChanBroker[K comparable, R any] struct {
	concurrency     int
	jobsChannel     chan K
	internalChan    chan *ChanResult[K, R]
	resultChan      chan *ChanResult[K, R]
	callback        func(K) R
	queueOfRequests *safeListStruct[K]
	arrayOfResults  []*ChanResult[K, R]
	waitGroup       *sync.WaitGroup
	finished        bool
}

func CreateChanBroker[K comparable, R any](concurrency int, processFunction func(K) R) *ChanBroker[K, R] {
	buffLen := concurrency * 2
	arrayOfResults := make([]*ChanResult[K, R], 0, buffLen)
	internalChan := make(chan *ChanResult[K, R])
	jobsChannel := make(chan K, concurrency*2)
	queueOfRequests := MakeSafeList[K](concurrency * 2)
	resultChan := make(chan *ChanResult[K, R], concurrency*2)

	broker := &ChanBroker[K, R]{
		concurrency,
		jobsChannel,
		internalChan,
		resultChan,
		processFunction,
		&queueOfRequests,
		arrayOfResults,
		&sync.WaitGroup{},
		false,
	}

	broker.waitGroup.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go broker.spawnWorker()
	}

	go broker.start()

	return broker
}

func (broker *ChanBroker[K, R]) AddJob(job K) {
	broker.queueOfRequests.Push(job)
	broker.jobsChannel <- job
}

func (broker *ChanBroker[K, R]) GetResultsChan() chan *ChanResult[K, R] {
	return broker.resultChan
}

func (broker *ChanBroker[K, R]) Finalize() {
	close(broker.jobsChannel)
	broker.internalChan <- &ChanResult[K, R]{final: true}
}

// ----------------

func (broker *ChanBroker[K, R]) tryToDequeue() bool {
	firstElementValue, ok := broker.queueOfRequests.First()
	resultsNum := len(broker.arrayOfResults)

	if ok && resultsNum > 0 {
		for i := 0; i < resultsNum; i++ {
			result := broker.arrayOfResults[i]
			if result.Key == firstElementValue {
				broker.arrayOfResults = append(broker.arrayOfResults[:i], broker.arrayOfResults[i+1:]...)
				broker.queueOfRequests.Remove(firstElementValue)
				broker.resultChan <- result
				return true
			}
		}
	}
	return false
}

func (broker *ChanBroker[K, R]) spawnWorker() {
	defer (func() {
		broker.waitGroup.Done()
	})()

	for task := range broker.jobsChannel {
		broker.internalChan <- &ChanResult[K, R]{
			Key:    task,
			Result: broker.callback(task),
			final:  false,
		}
	}
}

func (broker *ChanBroker[K, R]) start() {
	defer (func() {
		close(broker.resultChan)
		if len(broker.arrayOfResults) > 0 {
			panic("Non all results were returned - please report to github.com/krikus/go-chan or upgrade to newer version")
		}
	})()

	for result := range broker.internalChan {
		if result.final == false {
			broker.arrayOfResults = append(broker.arrayOfResults, result)

			for broker.tryToDequeue() {
			}
		} else {
			go (func() {
				broker.waitGroup.Wait()
				close(broker.internalChan)
				broker.finished = true
			})()
		}
	}
}
