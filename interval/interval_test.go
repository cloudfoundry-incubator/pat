package interval_test

import (
	. "github.com/julz/pat/interval"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interval", func() {
	//backfill more tests

	var (
		testFunc func()
	)

	BeforeEach(func() {
		testFunc = func() {
			return
		}
	})

	Describe("Repeat returns nil or *ReturnItem depends on the value of second", func() {

		Context("When Repeat() is called with seconds and fn() provide", func() {

			It("Should return a pointer to RepeatItem", func() {
				var ptr *RepeatItem
				var second = 1
				Ω(Repeat(second, testFunc)).Should(BeAssignableToTypeOf(ptr))
			})

		})

	})
})
