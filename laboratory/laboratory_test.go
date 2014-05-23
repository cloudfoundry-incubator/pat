package laboratory

import (
	. "github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Laboratory", func() {
	Describe("Running an experiment", func() {
		var (
			lab             Laboratory
			store           *dummyStore
			experiment1     Runnable
			experiment2     Runnable
			run1            string
			run2            string
			handlerRecieved []*Sample
		)

		BeforeEach(func() {
			store = &dummyStore{make(map[string][]*Sample), make([]Experiment, 0)}
		})

		JustBeforeEach(func() {
			lab = NewLaboratory(store)
			experiment1 = &dummyExperiment{"1", []*Sample{&Sample{}}}
			experiment2 = &dummyExperiment{"2", []*Sample{&Sample{}, &Sample{}}}
			run1, _ = lab.Run(experiment1)

			handlerRecieved = nil
			run2, _ = lab.RunWithHandlers(experiment2, []func(<-chan *Sample){func(samples <-chan *Sample) {
				for s := range samples {
					handlerRecieved = append(handlerRecieved, s)
				}
			}})

			// wait for experiments to run
			Eventually(func() int {
				return len(store.stored)
			}).Should(Equal(2))
		})

		It("gives experiments a GUID", func() {
			Ω(run1).ShouldNot(Equal("1"))
			Ω(run1).ShouldNot(Equal(run2))
		})

		It("saves running experiments to the store", func() {
			Ω(store.stored[run1]).Should(HaveLen(3))
			Ω(store.stored[run2]).Should(HaveLen(3))
		})

		It("sends data from a running experiment to a handler function", func() {
			Ω(handlerRecieved).Should(HaveLen(3))
		})

		Describe("Loading previous experiment at startup", func() {
			var (
				loadedExperiment1 Experiment
				loadedExperiment2 Experiment
				loadedData1       []*Sample
				loadedData2       []*Sample
			)

			BeforeEach(func() {
				loadedData1 = []*Sample{&Sample{}, &Sample{}}
				loadedData2 = []*Sample{&Sample{}, &Sample{}, &Sample{}}
				loadedExperiment1 = &dummyExperiment{"load1", loadedData1}
				loadedExperiment2 = &dummyExperiment{"load2", loadedData2}
				store.previous = append(store.previous, loadedExperiment1)
				store.previous = append(store.previous, loadedExperiment2)
			})

			It("retrieves data from a loaded experiment (by calling GetData())", func() {
				Ω(data(lab.GetData("load1"))).Should(HaveLen(len(loadedData1)))
				Ω(data(lab.GetData("load2"))).Should(HaveLen(len(loadedData2)))
			})

			It("lists saved and running experiments", func() {
				var got []Experiment
				Eventually(func() []Experiment {
					got = make([]Experiment, 0)
					lab.Visit(func(e Experiment) {
						got = append(got, e)
					})
					return got
				}).Should(HaveLen(2))
				Ω(got[0]).Should(Equal(loadedExperiment1))
				Ω(got[1]).Should(Equal(loadedExperiment2))
			})

			Describe("Reloading experiments", func() {
				var (
					laterExperiment Experiment
					loadedData      []*Sample
				)

				JustBeforeEach(func() {
					loadedData = []*Sample{&Sample{}, &Sample{}}
					laterExperiment = &dummyExperiment{"later", loadedData}
					store.previous = append(store.previous, laterExperiment)

				})

				It("Loads new experiments from the store dynamically", func() {
					var got []Experiment
					Eventually(func() []Experiment {
						got = make([]Experiment, 0)
						lab.Visit(func(e Experiment) {
							got = append(got, e)
						})
						return got
					}).Should(HaveLen(3))

					Ω(got[2]).Should(Equal(laterExperiment))
					Ω(got[2].GetGuid()).Should(Equal(laterExperiment.GetGuid()))
				})

				It("Retrieves data from a reloaded experiment", func() {
					Ω(data(lab.GetData("later"))).Should(HaveLen(len(loadedData)))
				})
			})
		})
	})
})

func data(s []*Sample, e error) []*Sample {
	Ω(e).ShouldNot(HaveOccurred())
	return s
}

type dummyStore struct {
	stored   map[string][]*Sample
	previous []Experiment
}

type dummyExperiment struct {
	name string
	data []*Sample
}

func (store *dummyStore) Writer(meta map[string]string) func(samples <-chan *Sample) {
	return func(samples <-chan *Sample) {
		store.stored[meta["guid"]] = make([]*Sample, 0)
		for s := range samples {
			store.stored[meta["guid"]] = append(store.stored[meta["guid"]], s)
		}
	}
}

func (store *dummyStore) LoadAll() ([]Experiment, error) {
	return store.previous, nil
}

func (e *dummyExperiment) Run(fn func(samples <-chan *Sample)) error {
	ch := make(chan *Sample)
	done := make(chan bool)
	go func() {
		fn(ch)
		done <- true
	}()
	ch <- &Sample{}
	ch <- &Sample{}
	ch <- &Sample{}
	close(ch)
	<-done
	return nil
}

func (e *dummyExperiment) GetData() ([]*Sample, error) {
	return e.data, nil
}

func (e *dummyExperiment) GetGuid() string {
	return e.name
}

func (e *dummyExperiment) DescribeMetadata() map[string]string {
	mMap := NewExperimentConfiguration(0, nil, 0, 0, 0, nil, "", "").DescribeMetadata()
	return mMap
}
