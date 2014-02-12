package server_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/julz/pat/config"
	. "github.com/julz/pat/experiment"
	. "github.com/julz/pat/laboratory"
	. "github.com/julz/pat/server"
	"github.com/julz/pat/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var (
		lab *DummyLab
	)

	BeforeEach(func() {
		experiments := []*DummyExperiment{&DummyExperiment{"a"}, &DummyExperiment{"b"}, &DummyExperiment{"c"}}
		lab = &DummyLab{}
		lab.experiments = experiments
		http.DefaultServeMux = http.NewServeMux()
		ServeWithLab(lab)
	})

	It("Uses config to get CSV output directory", func() {
		http.DefaultServeMux = http.NewServeMux()
		c := config.NewConfig()
		InitCommandLineFlags(c)
		c.Parse([]string{"-csvDir", "/var/tmp/foo/bar/"})
		csvs := store.NewCsvStore("/var/tmp/foo/bar/")
		ch := make(chan *Sample)
		go func() { ch <- &Sample{}; ch <- &Sample{}; close(ch) }()
		csvs.Writer("1234")(ch)

		Serve()
		json := get("/experiments/1234")
		Ω(json["Items"]).Should(HaveLen(2))
	})

	It("lists experiments", func() {
		json := get("/experiments/")
		Ω(json["Items"]).Should(HaveLen(3))
		items := json["Items"].([]interface{})
		Ω(items[0].(map[string]interface{})["Location"]).Should(Equal("/experiments/a"))
		Ω(items[1].(map[string]interface{})["Location"]).Should(Equal("/experiments/b"))
		Ω(items[2].(map[string]interface{})["Location"]).Should(Equal("/experiments/c"))
	})

	It("lists experiments with a Csv Url link", func() {
		json := get("/experiments/")
		Ω(json["Items"]).Should(HaveLen(3))
		items := json["Items"].([]interface{})
		Ω(items[0].(map[string]interface{})["CsvLocation"]).Should(
			Equal("/experiments/a.csv"))
	})

	It("exports an experiment as a CSV", func() {
		csv := req("GET", "/experiments/a.csv")
		lines := strings.Split(string(csv), "\n")
		Ω(lines).Should(HaveLen(1 + 3 + 1)) // header, rows, newline
		Ω(lines[0]).Should(ContainSubstring("Average,TotalTime,Total"))
		Ω(lines[1]).Should(ContainSubstring("0,0,0"))
	})

	It("Runs experiment with default arguments", func() {
		post("/experiments/")
		Ω(lab.config.Iterations).Should(Equal(1))
		Ω(lab.config.Concurrency).Should(Equal(1))
		Ω(lab.config.Workload).Should(Equal("push"))
	})

	It("Supports an 'iterations' parameter", func() {
		post("/experiments/?iterations=3")
		Ω(lab.config.Iterations).Should(Equal(3))
	})

	It("Supports a 'concurrency' parameter", func() {
		post("/experiments/?concurrency=3")
		Ω(lab.config.Concurrency).Should(Equal(3))
	})

	It("Supports a 'workload' parameter", func() {
		post("/experiments/?workload=flibble")
		Ω(lab.config.Workload).Should(Equal("flibble"))
	})

	It("Returns Location based on assigned experiment GUID", func() {
		json := post("/experiments/")
		Ω(json["Location"]).Should(Equal("/experiments/some-guid"))
	})
})

type DummyLab struct {
	experiments []*DummyExperiment
	config      *RunnableExperiment
}

type DummyExperiment struct {
	guid string
}

func (l *DummyLab) RunWithHandlers(ex Runnable, fns []func(<-chan *Sample)) (Experiment, error) {
	Fail("called unexpected dummy function")
	return nil, nil
}

func (l *DummyLab) Run(ex Runnable) (Experiment, error) {
	l.config = ex.(*RunnableExperiment)
	return &DummyExperiment{"some-guid"}, nil
}

func (l *DummyLab) Visit(fn func(ex Experiment)) {
	for _, e := range l.experiments {
		fn(e)
	}
}

func (l *DummyLab) GetData(name string) ([]*Sample, error) {
	if name == "a" {
		return []*Sample{&Sample{}, &Sample{}, &Sample{}}, nil
	}
	return nil, nil
}

func (e *DummyExperiment) GetData() ([]*Sample, error) {
	return nil, nil
}

func (e *DummyExperiment) GetGuid() string {
	return e.guid
}

func post(url string) (json map[string]interface{}) {
	return decode(req("POST", url))
}

func get(url string) (json map[string]interface{}) {
	return decode(req("GET", url))
}

func decode(encoded []byte) (decoded map[string]interface{}) {
	json.Unmarshal(encoded, &decoded)
	return decoded
}

func req(method string, url string) []byte {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		Ω(err).NotTo(HaveOccured())
	}

	http.DefaultServeMux.ServeHTTP(resp, req)
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		Ω(err).NotTo(HaveOccured())
		return nil
	} else {
		fmt.Printf("Body: %s", body)
		return body
	}
}
