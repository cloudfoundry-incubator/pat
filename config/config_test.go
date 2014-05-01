package config_test

import (
	"io/ioutil"
	"os"

	. "github.com/cloudfoundry-incubator/pat/config"
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
			value  string
			value2 string
			flags  []string
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

			Describe("When it's bound twice", func() {
				It("does not allow double-binding unless the target is the same", func() {
					Ω(func() { config.StringVar(&value2, "name", "", "description") }).Should(Panic())
				})

				It("allows double-binding if the target is the same", func() {
					Ω(func() { config.StringVar(&value, "name", "", "description") }).ShouldNot(Panic())
				})
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
			value  int
			value2 int
			flags  []string
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

			Describe("When it's bound twice", func() {
				It("does not allow double-binding unless the target is the same", func() {
					Ω(func() { config.IntVar(&value2, "name", 1, "description") }).Should(Panic())
				})

				It("allows double-binding if the target is the same", func() {
					Ω(func() { config.IntVar(&value, "name", 1, "description") }).ShouldNot(Panic())
				})
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
			value  bool
			value2 bool
			flags  []string
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

			Describe("When it's bound twice", func() {
				It("does not allow double-binding unless the target is the same", func() {
					Ω(func() { config.BoolVar(&value2, "name", false, "description") }).Should(Panic())
				})

				It("allows double-binding if the target is the same", func() {
					Ω(func() { config.BoolVar(&value, "name", false, "description") }).ShouldNot(Panic())
				})
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

	Describe("Binding an environment variable", func() {
		var (
			value  string
			config Config
			flags  []string
		)

		BeforeEach(func() {
			config = NewConfig()
			flags = []string{}
			config.EnvVar(&value, "NAME", "a default value", "an environment variable")
			os.Clearenv()
		})

		JustBeforeEach(func() {
			config.Parse(flags)
		})

		Context("When the env variable is not set", func() {
			It("uses the default value", func() {
				Ω(value).Should(Equal("a default value"))
			})
		})

		Context("When the env variable is set", func() {
			BeforeEach(func() {
				os.Setenv("NAME", "the value")
			})

			It("uses the default value", func() {
				Ω(value).Should(Equal("the value"))
			})
		})
	})
})
