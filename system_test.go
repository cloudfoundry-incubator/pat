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
	"time"
)

var _ = Describe("System", func() {
	Describe("Running PATs with a web API", func() {
		BeforeEach(func() {
			os.RemoveAll("/tmp/pats-acceptance-test-runs")
		})

		It("Reports app push speed correctly", func() {
			json := post("/experiments/")
			Ω(json["Location"]).ShouldNot(BeNil())

			time.Sleep(1 * time.Second) // yuck- jz

			resp := get(json["Location"].(string))
			Ω(resp["Items"]).ShouldNot(BeEmpty())
		})

		PIt("Lists historical results", func() {
			// this test was flakey, going to replace
		})

		PIt("Lists results between two dates", func() {
			// this test was flakey, going to replace
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
