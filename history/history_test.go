package history_test

import (
  "encoding/json"
  "github.com/julz/pat/history"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "io/ioutil"
  "os"
  "path"
  "reflect"
)

type myResponse struct {
  Foo string
  Bar int
}

var _ = Describe("History", func() {
  var (
    reply  *myResponse
    result interface{}
    save   func()
    files  []os.FileInfo
  )

  BeforeEach(func() {
    os.RemoveAll("test-runs")
    reply = &myResponse{"foo", 2}

    save = func() {
      var err error
      result, err = history.Save("test-runs", reply)
      Expect(err).NotTo(HaveOccured())
      Expect(result).NotTo(BeNil())

      files, err = ioutil.ReadDir("test-runs")
      Expect(err).NotTo(HaveOccured())
    }
  })

  It("Saves the results to JSON", func() {
    save()

    Ω(files).Should(HaveLen(1))
    loaded, err := ioutil.ReadFile(path.Join("test-runs", files[0].Name()))
    Expect(err).NotTo(HaveOccured())

    var decoded myResponse
    err = json.Unmarshal(loaded, &decoded)
    Expect(err).NotTo(HaveOccured())

    Ω(result).Should(Equal(reply))
    Ω(decoded.Foo).Should(Equal("foo"))
    Ω(decoded.Bar).Should(Equal(2))
  })

  It("Saves each result in a new file", func() {
    save()
    save()

    Ω(files).Should(HaveLen(2))
  })

  It("Loads all the results back (pagination/date ranges not implemented yet)", func() {
    save()
    save()
    save()

    results, err := history.LoadAll("test-runs", reflect.TypeOf(myResponse{}))
    Ω(err).NotTo(HaveOccured())
    Ω(results).To(HaveLen(3))
    Ω(results[0].(*myResponse).Foo).Should(Equal("foo"))
  })
})
