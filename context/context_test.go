package context_test

import (
	"github.com/cloudfoundry-incubator/pat/context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context map", func() {

	var (
		localContext context.Context
	)

	JustBeforeEach(func() {
		localContext = context.New()
	})

	Context("String values in context map", func() {
		It("uses a key to identify store fields", func() {
			localContext.PutString("key1", "abc")
			localContext.PutString("key2", "123")

			result, exists := localContext.GetString("key1")
			Ω(result).Should(Equal("abc"))
			Ω(exists).Should(Equal(true))

			result, exists = localContext.GetString("key2")
			Ω(result).Should(Equal("123"))
			Ω(exists).Should(Equal(true))
		})

		It("can store string value as provided", func() {
			localContext.PutString("str", "This is a long string \n")

			result, _ := localContext.GetString("str")
			Ω(result).Should(Equal("This is a long string \n"))
		})

		It("can store int value as provided", func() {
			localContext.PutInt("int", 123)

			result, _ := localContext.GetInt("int")
			Ω(result).Should(Equal(123))
		})		

		It("can store bool value as provided", func() {
			localContext.PutBool("key", true)

			result, _ := localContext.GetBool("key")
			Ω(result).Should(Equal(true))
		})

		It("can store float64 value as provided", func() {
			localContext.PutFloat64("key", float64(3.14))

			result, _ := localContext.GetFloat64("key")
			Ω(result).Should(Equal(float64(3.14)))
		})
	})

	Context("Cloning map", func() {
		It("returns a copy of the cloned context map", func() {
			localContext.PutString("str1", "abc")
			localContext.PutString("str2", "def")

			cloned := localContext.Clone()

			localContext.PutString("str1", "123")

			result, _ := localContext.GetString("str1")
			Ω(result).Should(Equal("123"))

			result, _ = cloned.GetString("str1")
			Ω(result).Should(Equal("abc"))
		})
	})

})
