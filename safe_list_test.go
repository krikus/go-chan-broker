package gochanbroker

import (
	"math/rand"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	defer GinkgoRecover()
	list := MakeSafeList[string](10)

	When("List is empty", func() {
		When("Len() is called", func() {
			It("should return 0", func() {
				Expect(list.Len()).To(Equal(0))
			})
		})

		When("First() is called", func() {
			It("should return false", func() {
				_, ok := list.First()
				Expect(ok).To(Equal(false))
			})
		})

		When("Remove() is called", func() {
			It("should return false", func() {
				Expect(list.Remove("")).To(Equal(false))
			})
		})

		When("Pop() is called", func() {
			It("should return false", func() {
				_, ok := list.Pop()
				Expect(ok).To(Equal(false))
			})
		})
	})

	// Adding items
	When("List with two items", func() {
		list := MakeSafeList[string](0)
		list.Push("test1")
		list.Push("test2")

		When("Len() is called", func() {
			It("should return 2", func() {
				Expect(list.Len()).To(Equal(2))
			})
		})

		When("First() is called", func() {
			It("should return true", func() {
				_, ok := list.First()
				Expect(ok).To(Equal(true))
			})

			It("should return first item", func() {
				item, _ := list.First()
				Expect(item).To(Equal("test1"))
			})
		})

		When("Remove() is called", func() {
			defer GinkgoRecover()
			list := MakeSafeList[string](0)
			list.Push("test1")
			list.Push("test2")
			ok := list.Remove("test1")

			It("should return true", func() {
				Expect(ok).To(Equal(true))
			})

			It("Len() should return 1", func() {
				Expect(list.Len()).To(Equal(1))
			})
		})

		When("Removing more items", func() {
			list := MakeSafeList[string](0)
			list.Push("test1")
			list.Push("test2")
			list.Push("test3")
			ok := list.Remove("test2")
			first, firstOk := list.First()
			pop, popOk := list.Pop()

			It("Remove should return true", func() {
				Expect(ok).To(Equal(true))
			})

			It("First() should now return second item", func() {
				Expect(firstOk).To(Equal(true))
				Expect(first).To(Equal("test1"))
			})

			It("Pop() should return last item", func() {
				Expect(popOk).To(Equal(true))
				Expect(pop).To(Equal("test3"))
			})
		})
	})

})

var _ = Describe("Middle element handling", func() {
	defer GinkgoRecover()
	var list safeListStruct[string]
	list = MakeSafeList[string](10)
	list.Push("test1")
	list.Push("test2")
	list.Push("test3")
	list.Push("test4")
	list.Push("test5")

	When("Len() is called", func() {
		It("Should have length of 5", func() {
			Expect(list.Len()).To(Equal(5))
		})
	})

	When("Emptying list with Pop() call", func() {
		defer GinkgoRecover()
		list := list.Duplicate()
		items := []string{}
		for {
			item, ok := list.Pop()
			if !ok {
				break
			}
			items = append(items, item)
		}

		It("should return items in correct order", func() {
			Expect(items).To(Equal([]string{"test5", "test4", "test3", "test2", "test1"}))
		})

		It("List should be empty", func() {
			Expect(list.Len()).To(Equal(0))
		})
	})

	When("Removing middle element", func() {
		defer GinkgoRecover()
		list := list.Duplicate()
		ok := list.Remove("test3")

		It("should return true", func() {
			Expect(ok).To(Equal(true))
		})

		It("should have length of 4", func() {
			Expect(list.Len()).To(Equal(4))
		})

		When("Removing same element again", func() {
			ok := list.Remove("test3")
			It("should return false", func() {
				Expect(ok).To(Equal(false))
			})
		})
	})

})

var _ = Describe("Adding more items than buffor length in goroutines", func() {
	defer GinkgoRecover()
	rand.Seed(GinkgoRandomSeed())
	testLen := 1000
	middle := testLen / 2
	type testStruct struct {
		val int
	}
	consistArray := make([]testStruct, testLen)
	list := MakeSafeList[testStruct](testLen / 10)
	wait := sync.WaitGroup{}
	wait.Add(testLen)

	for i := 0; i < testLen; i++ {
		consistArray[i] = testStruct{val: i}
		go func(id int) {
			ms := rand.Intn(10)
			<-time.After(time.Duration(ms) * time.Millisecond)
			list.Push(consistArray[id])
			wait.Done()
		}(i)
	}

	wait.Wait()

	When("List is build by goroutines", func() {
		defer GinkgoRecover()
		It("should have all elements", func() {
			Expect(list.Len()).To(Equal(testLen))
		})
	})

	When("Removing middle element and edge ones", func() {
		defer GinkgoRecover()
		list := list.Duplicate()

		list.Remove(testStruct{val: middle})
		list.Remove(testStruct{val: 0})
		list.Remove(testStruct{val: testLen - 1})

		It("should have correct length", func() {
			Expect(list.Len()).To(Equal(testLen - 3))
		})
	})

	When("Emptying list with Pop() call after removing some elements", func() {
		defer GinkgoRecover()
		list := list.Duplicate()

		// removing middle one
		list.Remove(testStruct{val: middle})
		// removing first
		list.Remove(testStruct{val: 0})
		// removing last
		list.Remove(testStruct{val: testLen - 1})
		items := []testStruct{}
		for {
			item, ok := list.Pop()
			if !ok {
				break
			}
			items = append(items, item)
		}

		It("should have all elements besides middle and edges", func() {
			targetArray := append(consistArray[1:middle], consistArray[middle+1:testLen-1]...)
			Expect(items).To(ConsistOf(targetArray))
		})

		It("Len() should return 0", func() {
			Expect(list.Len()).To(Equal(0))
		})
	})
})
