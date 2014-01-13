package pat

import (
  "encoding/json"
  "fmt"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "io/ioutil"
  "net/http"
  "net/http/httptest"
)

var _ = Describe("System", func() {
	Describe("Running PATs with a cmd line interface", func() {
		It("Runs a push and responds with the speed", func() {
			output := RunCommandLine()
			Ω(output.TotalTime).Should(BeNumerically("~", 250000, 100000))
		})
	})

	Describe("Running PATs with a web API", func() {
		It("Reports app push speed correctly", func() {
			json := post("/experiments/push")
			Ω(json["totalTime"]).Should(BeNumerically("~", 250000, 100000))
		})
	})
})


func decode(encoded []byte) (decoded map[string]interface{}) {
  json.Unmarshal(encoded, &decoded)
  return decoded
}

func post(url string) (json map[string]interface{}) {
  resp := httptest.NewRecorder()
  req, err := http.NewRequest("POST", url, nil)
  if err != nil {
    Fail("Error creating POST request")
  }

  http.DefaultServeMux.ServeHTTP(resp, req)
  if body, err := ioutil.ReadAll(resp.Body); err != nil {
    Fail("Error POSTing to URL")
		return nil
	} else {
		fmt.Printf("Body: %s", body)
		return decode(body)
	}
}
