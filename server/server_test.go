package server_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/context"
	. "github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/cloudfoundry-incubator/pat/laboratory"
	. "github.com/cloudfoundry-incubator/pat/server"
	"github.com/cloudfoundry-incubator/pat/store"
	"github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var workloadContext context.Context

var _ = Describe("Server", func() {
	var (
		lab             *DummyLab
		workloadContext context.Context
	)

	BeforeEach(func() {
		experiments := []*DummyExperiment{&DummyExperiment{"a"}, &DummyExperiment{"b"}, &DummyExperiment{"c"}}
		lab = &DummyLab{}
		lab.experiments = experiments
		http.DefaultServeMux = http.NewServeMux()
		workloadContext = context.New()
	})

	JustBeforeEach(func() {
		ServeWithLab(lab)
	})

	Describe("VCAP_APP_PORT", func() {
		var (
			listen string
			flags  config.Config
		)

		BeforeEach(func() {
			flags = config.NewConfig()
			os.Clearenv()
			InitCommandLineFlags(flags)
			ListenAndServe = func(bind string) error {
				listen = bind
				return nil
			}
		})

		Context("When VCAP_APP_PORT does not exist", func() {
			BeforeEach(func() {
				os.Clearenv()
				flags.Parse([]string{})
			})

			It("Defaults to 8080", func() {
				Ω(listen).Should(Equal(":8080"))
			})
		})

		Context("When VCAP_APP_PORT exists", func() {
			BeforeEach(func() {
				os.Setenv("VCAP_APP_PORT", "1234")
				flags.Parse([]string{})
			})

			It("Uses the env variable", func() {
				Ω(listen).Should(Equal(":1234"))
			})
		})
	})

	It("Uses config to get CSV output directory", func() {
		http.DefaultServeMux = http.NewServeMux()
		c := config.NewConfig()
		tempDir := filepath.Join(os.TempDir(), "foo")
		InitCommandLineFlags(c)
		c.Parse([]string{"-csv-dir", tempDir})
		csvs := store.NewCsvStore(tempDir, workloads.DefaultWorkloadList())
		ch := make(chan *Sample)
		go func() { ch <- &Sample{}; ch <- &Sample{}; close(ch) }()
		csvs.Writer("1234")(ch)
		Serve()
		json := get("/experiments/1234")
		Ω(json["Items"]).Should(HaveLen(2))
		os.RemoveAll(tempDir)
	})

	It("Returns empty experiments as [] not null", func() {
		json := get("/experiments/empty")
		Ω(json["Items"]).ShouldNot(BeNil())
	})

	It("lists experiments", func() {
		json := get("/experiments/")
		Ω(json["Items"]).Should(HaveLen(3))
		items := json["Items"].([]interface{})
		Ω(items[0].(map[string]interface{})["Location"]).Should(Equal("/experiments/a"))
		Ω(items[1].(map[string]interface{})["Location"]).Should(Equal("/experiments/b"))
		Ω(items[2].(map[string]interface{})["Location"]).Should(Equal("/experiments/c"))
	})

	It("populates an experiment", func() {
		json := get("/experiments/a")
		Ω(json).Should(HaveLen(1))
		experimentA := json["Items"].([]interface{})[0]
		keys := []string{"Average", "Commands", "LastError", "NinetyfifthPercentile", "Total", "TotalTime", "TotalWorkers", "WallTime", "WorstResult", "LastResult", "TotalErrors", "Type"}
		for _, key := range keys {
			Ω(experimentA).Should(HaveKey(key))
		}
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
		Ω(lab.config.Concurrency).Should(Equal([]int{1}))
		Ω(lab.config.ConcurrencyStepTime).Should(Equal(60 * time.Second))
		Ω(lab.config.Interval).Should(Equal(0))
		Ω(lab.config.Stop).Should(Equal(0))
		Ω(lab.config.Workload).Should(Equal("gcf:push"))
	})

	It("Supports an 'iterations' parameter", func() {
		post("/experiments/?iterations=3")
		Ω(lab.config.Iterations).Should(Equal(3))
	})

	It("Supports a 'concurrency' parameter", func() {
		post("/experiments/?concurrency=3")
		Ω(lab.config.Concurrency).Should(Equal([]int{3}))
	})

	It("Supports a 'concurrency:timeBetweenSteps' parameter in seconds", func() {
		post("/experiments/?concurrency:timeBetweenSteps=3")
		Ω(lab.config.ConcurrencyStepTime).Should(Equal(3 * time.Second))
	})

	It("Supports an 'interval' parameter", func() {
		post("/experiments/?interval=3")
		Ω(lab.config.Interval).Should(Equal(3))
	})

	It("Supports a 'stop' parameter", func() {
		post("/experiments/?stop=3")
		Ω(lab.config.Stop).Should(Equal(3))
	})

	It("Supports a 'workload' parameter", func() {
		post("/experiments/?workload=flibble")
		Ω(lab.config.Workload).Should(Equal("flibble"))
	})

	It("Supports a 'cfTarget' parameter", func() {
		post("/experiments/?cfTarget=http://api.127.0.0.1")
		Ω(workloadCtxStringValue("rest:target")).Should(Equal("http://api.127.0.0.1"))
	})

	It("Supports a 'cfUsername' parameter", func() {
		post("/experiments/?cfUsername=user1")
		Ω(workloadCtxStringValue("rest:username")).Should(Equal("user1"))
	})

	It("Supports a 'cfPassword' parameter", func() {
		post("/experiments/?cfPassword=pass1")
		Ω(workloadCtxStringValue("rest:password")).Should(Equal("pass1"))
	})

	It("Supports a 'cfSpace' parameter", func() {
		post("/experiments/?cfSpace=dev123")
		Ω(workloadCtxStringValue("rest:space")).Should(Equal("dev123"))
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

func (l *DummyLab) RunWithHandlers(ex Runnable, fns []func(<-chan *Sample), workloadCtx context.Context) (string, error) {
	Fail("called unexpected dummy function")
	return "", nil
}

func (l *DummyLab) Run(ex Runnable, workloadCtx context.Context) (string, error) {
	l.config = ex.(*RunnableExperiment)
	workloadContext = workloadCtx
	return "some-guid", nil
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
		Ω(err).NotTo(HaveOccurred())
	}

	http.DefaultServeMux.ServeHTTP(resp, req)
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		Ω(err).NotTo(HaveOccurred())
		return nil
	} else {
		return body
	}
}

func workloadCtxStringValue(key string) string {
	str, _ := workloadContext.GetString(key)
	return str
}
