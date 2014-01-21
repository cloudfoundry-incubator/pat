package interval_test

import (
	. "github.com/julz/pat/interval"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interval", func() {
	//backfill more tests
	
	Describe("Repeat returns nil or *ReturnItem depends on the value of second", func() {
		testFunc := func() {
			return
		}
		
		Context("returns nil when seconds <= 0", func() {
			var second = 0
				Ω(Repeat(second, testFunc)).Should(BeNil())
		})

		Context("returns *RepeatItem when seconds > 0", func() {
			var ptr *RepeatItem
			var second = 1
			Ω(Repeat(second, testFunc)).Should(BeAssignableToTypeOf(ptr))
		})

	})
})
