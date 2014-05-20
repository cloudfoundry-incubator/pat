package helpers_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/pat/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File", func() {
	Context("#ReadWriteFile", func() {
		const (
			fileName = "tmp"
		)

		AfterEach(func() {
			os.Remove(fileName)
		})

		It("should return an error if the file cannot be created", func() {
			_, err := OpenOrCreate("")
			Ω(err).ShouldNot(BeNil())
		})

		It("should return the file if it exists", func() {
			_, err := os.Create(fileName)
			Ω(err).Should(BeNil())

			file, err := OpenOrCreate(fileName)
			Ω(file).ShouldNot(BeNil())
			Ω(err).Should(BeNil())
		})

		It("should create the file if it does not exist", func() {
			file, err := OpenOrCreate(fileName)
			Ω(file).ShouldNot(BeNil())
			Ω(err).Should(BeNil())
		})
	})

})
