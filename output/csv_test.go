package output_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/julz/pat/output"
	"github.com/julz/pat/experiment"
	"io/ioutil"
	"strings"
)

var _ = Describe("Csv", func() {
	It("Converts a list of experiments to a CSV", func() {
		filename := "/var/tmp/output/foo.csv"
		out := NewCsvWriter(filename)
		ch := make(chan *experiment.Sample)
		go func() {
			ch <- &experiment.Sample{}
			ch <- &experiment.Sample{}
			close(ch)
		}()
		out.Write(ch)

		in, err := ioutil.ReadFile(filename)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(strings.Split(string(in), "\n")[0]).Should(Equal("duration,wallTime,average,workers"))
		Ω(strings.Split(string(in), "\n")[1]).Should(Equal("0,0,0,0"))
		Ω(strings.Split(string(in), "\n")[2]).Should(Equal("0,0,0,0"))
	})
})
