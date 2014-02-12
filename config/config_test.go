package config_test

import (
	"io/ioutil"

	. "github.com/julz/pat/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const fileName = "TestConfigCmdFile.yml"

var _ = Describe("ConfigAndFlags", func() {
	var (
		config Config
	)

	BeforeEach(func() {
		config = NewConfig()
	})

	AfterEach(func() {
	})

	Describe("Adding a String flag", func() {
		var (
			value string
			flags []string
		)

		BeforeEach(func() {
			config.StringVar(&value, "name", "", "description")
		})

		JustBeforeEach(func() {
			config.Parse(flags)
		})

		Describe("When the parameter is provided as a flag", func() {
			BeforeEach(func() {
				flags = []string{"-name", "beans"}
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal("beans"))
			})
		})

		Describe("When the parameter is provided in a config file", func() {
			BeforeEach(func() {
				flags = []string{"-config", "/tmp/config.yml"}
				ioutil.WriteFile("/tmp/config.yml", []byte("name: branflakes"), 0755)
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal("branflakes"))
			})
		})

		Describe("When the parameter is provided in both config file and flag", func() {
			BeforeEach(func() {
				ioutil.WriteFile("/tmp/config.yml", []byte("name: branflakes"), 0755)
				flags = []string{"-config", "/tmp/config.yml", "-name", "beans"}
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal("beans"))
			})
		})
	})

	Describe("Adding an Integer flag", func() {
		var (
			value int
			flags []string
		)

		BeforeEach(func() {
			config.IntVar(&value, "name", 8, "description")
		})

		JustBeforeEach(func() {
			config.Parse(flags)
		})

		Describe("When the parameter is provided as a flag", func() {
			BeforeEach(func() {
				flags = []string{"-name", "7"}
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal(7))
			})
		})

		Describe("When the parameter is provided in a config file", func() {
			BeforeEach(func() {
				flags = []string{"-config", "/tmp/config.yml"}
				ioutil.WriteFile("/tmp/config.yml", []byte("name: 5"), 0755)
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal(5))
			})
		})
	})

	Describe("Adding a Bool flag", func() {
		var (
			value bool
			flags []string
		)

		BeforeEach(func() {
			config.BoolVar(&value, "name", false, "description")
		})

		JustBeforeEach(func() {
			config.Parse(flags)
		})

		Describe("When the parameter is provided as a flag", func() {
			BeforeEach(func() {
				flags = []string{"-name"}
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal(true))
			})
		})

		Describe("When the parameter is provided in a config file", func() {
			BeforeEach(func() {
				flags = []string{"-config", "/tmp/config.yml"}
				ioutil.WriteFile("/tmp/config.yml", []byte("name: true"), 0755)
			})

			It("Reads the version from the flag", func() {
				Ω(value).Should(Equal(true))
			})
		})
	})
})
