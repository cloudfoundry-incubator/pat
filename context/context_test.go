package context

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)


var _ = Describe("Context map", func() {

	var (
		content WorkloadContent
		localContext WorkloadContext
	)

	JustBeforeEach(func() {
		content = NewWorkloadContent()
		localContext = WorkloadContext(content)
	})

	Context("String values in context map", func(){
		It("uses a key to identify store fields", func(){
			localContext.PutString("key1", "abc")
			localContext.PutString("key2", "123")

			Ω(localContext.GetString("key1")).Should(Equal("abc"))
			Ω(localContext.GetString("key2")).Should(Equal("123"))
		})
		
		It("can store string value as provided", func(){
			localContext.PutString("str", "This is a long string \n")

			Ω(localContext.GetString("str")).Should(Equal("This is a long string \n"))
		})

		It("can store int value as provided", func(){
			localContext.PutInt("int", 123)

			Ω(localContext.GetInt("int")).Should(Equal(123))
		})

		It("can store int64 value as provided", func(){
			localContext.PutInt64("key", int64(10000))

			Ω(localContext.GetInt64("key")).Should(Equal(int64(10000)))
		})

		It("can store bool value as provided", func(){
			localContext.PutBool("key", true)

			Ω(localContext.GetBool("key")).Should(Equal(true))
		})

		It("can store float64 value as provided", func(){
			localContext.PutFloat64("key", float64(3.14))

			Ω(localContext.GetFloat64("key")).Should(Equal(float64(3.14)))
		})
	})

	Context("Checking existing fields", func(){
		It("returns true for existing fields, and false for non-existing", func(){
			localContext.PutString("key1", "some string")

			Ω(localContext.CheckExists("key1")).Should(Equal(true))
			Ω(localContext.CheckExists("key2")).Should(Equal(false))
		})
	})

	Context("Checking Types", func(){
		It("returns the type name of the field", func(){
			localContext.PutFloat64("key", float64(3.14))

			Ω(localContext.CheckType("key")).Should(Equal("float64"))
		})
	})

	Context("Cloning map", func(){
		It("returns a copy of the cloned context map", func(){
			localContext.PutString("str1", "abc")
			localContext.PutString("str2", "def")

			cloned := localContext.Clone()

			localContext.PutString("str1", "123")

			Ω(localContext.GetString("str1")).Should(Equal("123"))
			Ω(cloned.GetString("str1")).Should(Equal("abc"))
		})
	})

	Context("Getting map keys", func(){
		It("returns an array of key names", func(){
			localContext.PutString("key", "some value")
			localContext.PutString("key1", "some value")
			localContext.PutString("key2", "some value")
			localContext.PutString("key3", "some value")

			Ω(localContext.GetKeys()).Should(Equal([]string{"key","key1","key2","key3"}))
		})
	})

})