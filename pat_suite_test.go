package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPat(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pat Suite")
}
