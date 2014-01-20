package parser_test

import (
	. "github.com/julz/pat/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const fName = "workload.yml"

var _ = Describe("Workload", func() {
	It("Can Load in a yaml file", func() {
		work, err := ParseWorkload(fName)

		Ω(err).Should(BeNill)
	})
})
