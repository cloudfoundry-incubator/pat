package store_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/cloudfoundry-incubator/pat/store"
	"github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Csv Store", func() {
	const (
		dir = "/var/tmp/test-output/csvstore"
	)

	var (
		store            *CsvStore
		experimentConfig experiment.ExperimentConfiguration
		output           string
	)

	Describe("CsvStore", func() {
		It("Returns a CsvFile named after the experiment", func() {

		})
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	Describe("CsvFile", func() {
		var (
			commands map[string]experiment.Command
		)

		JustBeforeEach(func() {
			os.RemoveAll(dir)

			testList := []workloads.WorkloadStep{
				workloads.Step("boo", func() error { return nil }, "a"),
			}
			store = NewCsvStore(dir, &workloads.WorkloadList{testList})
			writer := store.Writer(experiment.ExperimentConfiguration{Guid: "foo"}.DescribeMetadata())
			commands = make(map[string]experiment.Command)
			cmd := experiment.Command{1, 0.5, 2, 3, 4, 5}
			commands["boo"] = cmd
			write(writer, []*experiment.Sample{
				&experiment.Sample{commands, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{commands, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 7, 2, experiment.ResultSample},
			})
			files, err := ioutil.ReadDir(dir)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(files).Should(HaveLen(2)) //one csv file and one meta file
			in, err := ioutil.ReadFile(path.Join(dir, files[0].Name()))
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
			ex, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			samples, err := ex[0].GetData()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(samples[0]).Should(Equal(&experiment.Sample{commands, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample}))
		})

		It("Does not save error text, to avoid huge files", func() {
			ex, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			samples, err := ex[0].GetData()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(samples[0].LastError).Should(BeNil())
		})

		It("Loads multiple CSVs from a directory, in order", func() {
			foo := store.Writer(experiment.ExperimentConfiguration{Guid: "bar"}.DescribeMetadata())
			write(foo, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 3, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 7, 2, experiment.ResultSample},
			})

			bar := store.Writer(experiment.ExperimentConfiguration{Guid: "baz"}.DescribeMetadata())
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

		It("Does not load load meta data inforomation", func() {
			samples, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(samples).Should(HaveLen(1))

			Ω(strings.Split(output, "\n")[0]).ShouldNot(ContainSubstring("csv guid"))
		})

		PIt("Throws exception if header is not in correct order", func() {})

		PIt("Saves a full and partial version with ErrorSample etc.", func() {})
	})

	Describe("MetaFile", func() {
		const (
			iterations          = 2
			concurrencyStepTime = time.Duration(5)
			interval            = 10
			stop                = 100
			workload            = "gcf:push"
			description         = "note description"
		)

		var (
			concurrency = []int{1, 2, 3}
		)

		Context("Creating a new file", func() {
			BeforeEach(func() {
				os.RemoveAll(dir)

				experimentConfig = experiment.NewExperimentConfiguration(
					iterations, concurrency, concurrencyStepTime,
					interval, stop, nil, workload, description)

				testList := []workloads.WorkloadStep{
					workloads.Step("boo", func() error { return nil }, "a"),
				}
				store = NewCsvStore(dir, &workloads.WorkloadList{testList})
				store.Writer(experimentConfig.DescribeMetadata())

				in, err := ioutil.ReadFile(path.Join(dir, "csv.meta"))
				Ω(err).ShouldNot(HaveOccurred())
				output = string(in)
			})

			It("saves the experiment's meta data headers as the first row in the meta file", func() {
				Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring(
					"csv guid,start time,iterations,concurrency,concurrency step time,stop,interval,workload,description"))
			})

			It("saves the guid of the experiment as the first item in the meta data", func() {
				data := strings.Split(output, "\n")[1]
				savedGuid := strings.Split(data, ",")[0]
				Ω(savedGuid).Should(Equal(experimentConfig.Guid))
			})

			It("saves the time of the experiment as the second item in the meta data", func() {
				data := strings.Split(output, "\n")[1]
				csvSplit := strings.Split(data, ",")
				t, err := time.Parse(time.RFC850, csvSplit[1][1:]+","+csvSplit[2][0:len(csvSplit[2])-1])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(t).Should(BeAssignableToTypeOf(time.Time{}))
			})

			It("saves the iteration meta data after time", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[3]).Should(Equal(strconv.Itoa(iterations)))
			})

			It("saves the concurrency meta data after iterations", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[4]).Should(Equal("1..2..3"))
			})

			It("saves the concurrency stop time meta data after concurrency", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[5]).Should(Equal(concurrencyStepTime.String()))
			})

			It("saves the stop meta data after concurency step time", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[6]).Should(Equal(strconv.Itoa(stop)))
			})

			It("saves the interval meta data after stop", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[7]).Should(Equal(strconv.Itoa(interval)))
			})

			It("saves the workload meta data after interval", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[8]).Should(Equal(workload))
			})

			It("saves the description meta data after workload", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[9]).Should(Equal(description))
			})
		})

		Context("When a meta data file already exists", func() {
			BeforeEach(func() {
				os.RemoveAll(dir)

				experimentConfig = experiment.NewExperimentConfiguration(
					iterations, concurrency, concurrencyStepTime,
					interval, stop, nil, workload, description)

				testList := []workloads.WorkloadStep{
					workloads.Step("boo", func() error { return nil }, "a"),
				}
				store = NewCsvStore(dir, &workloads.WorkloadList{testList})

				writer1 := store.Writer(experimentConfig.DescribeMetadata())
				writer2 := store.Writer(experimentConfig.DescribeMetadata())
				write(writer1, nil)
				write(writer2, nil)

				in, err := ioutil.ReadFile(path.Join(dir, "csv.meta"))
				Ω(err).ShouldNot(HaveOccurred())
				output = string(in)
			})

			It("saves the guid of the experiment as the first item in the meta data", func() {
				data := strings.Split(output, "\n")[2]
				savedGuid := strings.Split(data, ",")[0]
				Ω(savedGuid).Should(Equal(experimentConfig.Guid))
			})
		})
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
