package parser_test

import (
	"os"
	"io/ioutil"
	"launchpad.net/goyaml"
	. "github.com/julz/pat/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const fName = "TestConfigCmdFile.yml"

var _ = Describe("PatConfig", func() {
	var fStruct PATs

	BeforeEach(func() {
		fStruct = PATs{}
	})

	AfterEach(func() {
		deleteTestFile()
	})

	It("Can read in a configuration file with all peremeters set from a YAML file", func() {
		fStruct.Cli_commands.Server = true
		fStruct.Cli_commands.Pushes = 1
		fStruct.Cli_commands.Concurrency = 1
		fStruct.Cli_commands.Silent = true
		fStruct.Cli_commands.Output = "AFileName.csv"
		createTestFile(fStruct)

		pat, err := NewPATsConfiguration(fName)

		Ω(err).Should(BeNil())
		Ω(pat.Cli_commands.Server).Should(Equal(true))
		Ω(pat.Cli_commands.Pushes).Should(Equal(1))
		Ω(pat.Cli_commands.Concurrency).Should(Equal(1))
		Ω(pat.Cli_commands.Silent).Should(Equal(true))
		Ω(pat.Cli_commands.Output).Should(Equal("AFileName.csv"))
	})

	It("Can read in a configuration file with only some parameters set from a YAML file", func() {
		fStruct.Cli_commands.Server = true
		createTestFile(fStruct)

		pat, err := NewPATsConfiguration(fName)

		Ω(err).Should(BeNil())
		Ω(pat.Cli_commands.Server).Should(Equal(true))
	})

	It("Should return an error if the file cannot be found", func() {
		_, err := NewPATsConfiguration("")

		Ω(err).ShouldNot(BeNil())
	})

})

func createTestFile(T interface{}) {
	file, err := goyaml.Marshal(&T)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fName, file, 0644)
	if err != nil {
		panic(err)
	}
}

func deleteTestFile() {
	os.Remove(fName)
}
