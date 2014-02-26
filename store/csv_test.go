package store_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/julz/pat/experiment"
	. "github.com/julz/pat/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Csv Store", func() {
	Describe("CsvStore", func() {
		It("Returns a CsvFile named after the experiment", func() {

		})
	})

	Describe("CsvFile", func() {
		var (
			dir    string
			store  *CsvStore
			output string
		)

		JustBeforeEach(func() {
			dir = "/var/tmp/test-output/csvstore"
			os.RemoveAll(dir)
			store = NewCsvStore(dir)
			writer := store.Writer("foo")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 7, 2, experiment.ResultSample},
			})
			files, err := ioutil.ReadDir(dir)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(files).Should(HaveLen(1))
			in, err := ioutil.ReadFile(path.Join(dir, files[0].Name()))
			Ω(err).ShouldNot(HaveOccurred())
			output = string(in)
		})

		It("Converts a list of experiments to a CSV", func() {
			Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring("Average,TotalTime,Total,TotalErrors"))
			Ω(strings.Split(output, "\n")[1]).Should(ContainSubstring("1,2,3,4"))
			Ω(strings.Split(output, "\n")[2]).Should(ContainSubstring("9,8,7,6"))
		})

		It("Includes all fields, except LastError and Commands", func() {
			meta := reflect.ValueOf(experiment.Sample{}).Type()
			for i := 0; i < meta.NumField(); i++ {
				if meta.Field(i).Name == "Commands" {
					continue
				}

				if meta.Field(i).Name == "LastError" {
					continue
				}

				Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring(meta.Field(i).Name))
			}
		})

		It("Round trips", func() {
			ex, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			samples, err := ex[0].GetData()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(samples[0]).Should(Equal(&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample}))
		})

		It("Does not save error text, to avoid huge files", func() {
			ex, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			samples, err := ex[0].GetData()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(samples[0].LastError).Should(BeNil())
		})

		It("Loads multiple CSVs from a directory, in order", func() {
			foo := store.Writer("bar")
			write(foo, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 7, 2, experiment.ResultSample},
			})

			bar := store.Writer("baz")
			write(bar, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 7, 2, experiment.ResultSample},
			})

			samples, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(samples[0].GetGuid()).Should(Equal("foo"))
			Ω(samples[1].GetGuid()).Should(Equal("bar"))
			Ω(samples[2].GetGuid()).Should(Equal("baz"))
			Ω(data(samples[0].GetData())).Should(HaveLen(2))
			Ω(data(samples[1].GetData())).Should(HaveLen(2))
			Ω(data(samples[2].GetData())).Should(HaveLen(3))
		})

		PIt("Throws exception if header is not in correct order", func() {
		})

		PIt("Saves a full and partial version with ErrorSample etc.", func() {})
	})
})

func data(samples []*experiment.Sample, err error) []*experiment.Sample {
	Ω(err).ShouldNot(HaveOccurred())
	return samples
}

func write(writer func(samples <-chan *experiment.Sample), samples []*experiment.Sample) {
	ch := make(chan *experiment.Sample)
	go func() {
		for _, s := range samples {
			ch <- s
		}
		close(ch)
	}()
	writer(ch)
}

type dummyExperiment struct {
	name string
}

func (e *dummyExperiment) Run(handler func(samples <-chan *experiment.Sample)) error {
	return nil
}

func (e *dummyExperiment) GetGuid() string {
	return e.name
}

func (e *dummyExperiment) GetData() ([]*experiment.Sample, error) {
	return make([]*experiment.Sample, 0), nil
}
