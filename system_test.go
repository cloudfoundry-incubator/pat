package pat

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
)

var _ = Describe("System", func() {
	Describe("Running PATs with a cmd line interface", func() {
		It("Runs a push and responds with the speed", func() {
			output := RunCommandLine()
			Ω(output.TotalTime).Should(BeNumerically("~", 44088715219, 3000000000))
		})
	})

	Describe("Running PATs with a web API", func() {
		BeforeEach(func() {
			os.RemoveAll("/tmp/pats-acceptance-test-runs")
		})

		It("Reports app push speed correctly", func() {
			json := post("/experiments/push")
			Ω(json["TotalTime"]).Should(BeNumerically("~", 44088715219, 3000000000))
		})

		It("Lists historical results", func() {
			post("/experiments/push")
			post("/experiments/push")

			json := get("/experiments")
			Ω(json["Items"]).Should(HaveLen(2))
		})
	})
})

func decode(encoded []byte) (decoded map[string]interface{}) {
	json.Unmarshal(encoded, &decoded)
	return decoded
}

func post(url string) (json map[string]interface{}) {
	return req("POST", url)
}

func get(url string) (json map[string]interface{}) {
	return req("GET", url)
}

func req(method string, url string) (json map[string]interface{}) {
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
		return decode(body)
	}
}
