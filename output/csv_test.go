package output_test

import (
	"errors"
	"github.com/julz/pat/experiment"
	. "github.com/julz/pat/output"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"reflect"
	"strings"
)

var _ = Describe("Csv", func() {
	var (
		output string
		csv    *CsvSampleFile
	)

	JustBeforeEach(func() {
		filename := "/var/tmp/test-output/foo.csv"
		csv = write(filename, []*experiment.Sample{
			&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
			&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 2, experiment.ResultSample},
		})
		in, err := ioutil.ReadFile(filename)
		Ω(err).ShouldNot(HaveOccurred())
		output = string(in)
	})

	It("Converts a list of experiments to a CSV", func() {
		Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring("Average,TotalTime,Total,TotalErrors"))
		Ω(strings.Split(output, "\n")[1]).Should(ContainSubstring("1,2,3,4"))
		Ω(strings.Split(output, "\n")[2]).Should(ContainSubstring("9,8,7,6"))
	})

	It("Includes all fields, except LastError", func() {
		meta := reflect.ValueOf(experiment.Sample{}).Type()
		for i := 0; i < meta.NumField(); i++ {
			if meta.Field(i).Name == "LastError" {
				continue
			}

			Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring(meta.Field(i).Name))
		}
	})

	It("Round trips", func() {
		samples, err := csv.Read()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(samples[0]).Should(Equal(&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample}))
	})

	It("Does not save error text, to avoid huge files", func() {
		samples, err := csv.Read()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(samples[0].LastError).Should(BeNil())
	})

	It("Loads multiple CSVs from a directory", func() {
		write("/var/tmp/test-output/foo.csv", []*experiment.Sample{
			&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
			&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 2, experiment.ResultSample},
		})

		write("/var/tmp/test-output/bar.csv", []*experiment.Sample{
			&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
			&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 2, experiment.ResultSample},
		})

		samples, err := ReloadCSVs("/var/tmp/test-output/")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(samples["foo"]).Should(HaveLen(2))
		Ω(samples["bar"]).Should(HaveLen(2))
	})

	PIt("Throws exception if header is not in correct order", func() {
	})

	PIt("Saves a full and partial version with ErrorSample etc.", func() {})
})

func write(filename string, samples []*experiment.Sample) *CsvSampleFile {
	csv := NewCsvWriter(filename)
	ch := make(chan *experiment.Sample)
	go func() {
		for _, s := range samples {
			ch <- s
		}
		close(ch)
	}()
	csv.Write(ch)

	return csv
}
