package logs

import (
	"os"

	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry/gosteno"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var (
		args       []string
		sink       *notASink
		flags      config.Config
		calledInit int
	)

	BeforeEach(func() {
		initialized = false
		sink = &notASink{}
		args = []string{}
		flags = config.NewConfig()

		NewIOSink = func(target *os.File) gosteno.Sink {
			sink.target = target
			sink.sinkType = "io"
			return sink
		}

		NewFileSink = func(target string) gosteno.Sink {
			sink.target = target
			sink.sinkType = "file"
			return sink
		}

		InitGoSteno = func(c *gosteno.Config) {
			calledInit++
			gosteno.Init(c)
		}
	})

	JustBeforeEach(func() {
		InitCommandLineFlags(flags)
		flags.Parse(args)
	})

	It("Only calls gosteno.Init once", func() {
		NewLogger("abc").Info("easy as 123")
		NewLogger("so simple as").Info("abc, 123, you and me, baby")
		Ω(calledInit).Should(Equal(1))
	})

	It("Uses a JSON codec", func() {
		NewLogger("jenny.from.the.block")
		Ω(sink.GetCodec()).Should(BeAssignableToTypeOf(&gosteno.JsonCodec{}))
	})

	Context("When -logging:file is not specified", func() {
		It("logs to stdout", func() {
			NewLogger("jenny.from.the.block").Info("on the 6")
			Ω(sink.sinkType).Should(Equal("io"))
			Ω(sink.target).Should(Equal(os.Stdout))
		})
	})

	Context("When -logging:file is specified", func() {
		BeforeEach(func() {
			args = []string{"-logging:file", "/tmp/thelog"}
		})

		It("logs to the given file", func() {
			NewLogger("innerversions")
			Ω(sink.sinkType).Should(Equal("file"))
			Ω(sink.target).Should(Equal("/tmp/thelog"))
		})
	})

	Context("When -logging:level is empty", func() {
		BeforeEach(func() {
			args = []string{"-logging:level", ""}
		})

		It("Turns logs off (useful for tests)", func() {
			NewLogger("test").Debug("Debug")
			NewLogger("test").Info("Info")
			Ω(sink.records).Should(HaveLen(0))
		})
	})

	Context("When -logging:level is specified", func() {
		BeforeEach(func() {
			args = []string{"-logging:level", "debug"}
		})

		It("Logs at the given level", func() {
			NewLogger("test").Debug("Message")
			Ω(sink.records).Should(ContainElement("Message"))
		})

		Context(".. with capitalization errors", func() {
			BeforeEach(func() {
				args = []string{"-logging:level", "DEBUG"}
			})

			It("Still Logs at the given level", func() {
				NewLogger("test").Debug("Message")
				Ω(sink.records).Should(ContainElement("Message"))
			})
		})

		Context(".. but if the logging level is invalid", func() {
			BeforeEach(func() {
				args = []string{"-logging:level", "NotALoggingLevel"}
			})

			It("panicks", func() {
				Ω(func() { NewLogger("test") }).Should(Panic())
			})
		})
	})
})

type notASink struct {
	codec    gosteno.Codec
	records  []string
	target   interface{}
	sinkType string
}

func (s *notASink) AddRecord(record *gosteno.Record) {
	s.records = append(s.records, record.Message)
}

func (s *notASink) Flush() {
}

func (s *notASink) SetCodec(c gosteno.Codec) {
	s.codec = c
}

func (s *notASink) GetCodec() gosteno.Codec {
	return s.codec
}
