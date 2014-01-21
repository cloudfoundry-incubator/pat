package config_test

import (
	"flag"
	. "github.com/julz/pat/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
)

const fileName = "TestConfigCmdFile.yml"

var _ = Describe("PatConfig", func() {
	var (
		config = NewConfig()
		cfg    = Config{}
	)

	BeforeEach(func() {
		cfg = Config{}
		flag.Set("config", fileName)
	})

	AfterEach(func() {
		deleteTestFile()
	})

	It("Can read in a configuration file with all peremeters set from a YAML file", func() {
		cfg.Server = true
		cfg.Iterations = 1
		cfg.Concurrency = 1
		cfg.Silent = true
		cfg.Output = "AFileName.csv"
		cfg.Workload = "push,push"
		cfg.Interval = 5
		cfg.Stop = 5
		createTestFile(cfg)

		err := config.Parse()

		Ω(err).Should(BeNil())
		Ω(config.Server).Should(Equal(true))
		Ω(config.Iterations).Should(Equal(1))
		Ω(config.Concurrency).Should(Equal(1))
		Ω(config.Silent).Should(Equal(true))
		Ω(config.Output).Should(Equal("AFileName.csv"))
		Ω(config.Workload).Should(Equal("push,push"))
		Ω(config.Interval).Should(Equal(5))
		Ω(config.Stop).Should(Equal(5))
	})

	It("Can read in a configuration file with only some parameters set from a YAML file", func() {
		cfg.Server = true
		createTestFile(cfg)

		err := config.Parse()

		Ω(err).Should(BeNil())
		Ω(config.Server).Should(Equal(true))
		Ω(config.Silent).Should(Equal(false))
	})

	It("Can override a configuration file parameter if value is passed in by command line", func() {
		cfg.Iterations = 5
		createTestFile(cfg)

		err := config.Parse()
		flag.Set("iterations", "4")

		Ω(err).Should(BeNil())
		Ω(config.Iterations).Should(Equal(4))
	})

	It("Should return an error if the file cannot be found", func() {
		err := config.Parse()

		Ω(err).ShouldNot(BeNil())
	})
})

func createTestFile(T interface{}) {
	file, err := goyaml.Marshal(&T)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileName, file, 0644)
	if err != nil {
		panic(err)
	}
}

func deleteTestFile() {
	os.Remove(fileName)
}
