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
	"time"
)

type myResponse struct {
	Foo       string
	Bar       int
	Timestamp time.Time
}

var _ = Describe("History", func() {
	var (
		result interface{}
		save   func() *myResponse
		files  []os.FileInfo
	)

	BeforeEach(func() {
		os.RemoveAll("test-runs")

		save = func() *myResponse {
			var err error

			reply := &myResponse{"foo", 2, time.Now()}
			result, err = history.Save("test-runs", reply, reply.Timestamp.UnixNano())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal(reply))

			files, err = ioutil.ReadDir("test-runs")
			Expect(err).NotTo(HaveOccurred())
			return reply
		}
	})

	It("Saves the results to JSON", func() {
		save()

		Ω(files).Should(HaveLen(1))
		loaded, err := ioutil.ReadFile(path.Join("test-runs", files[0].Name()))
		Expect(err).NotTo(HaveOccurred())

		var decoded myResponse
		err = json.Unmarshal(loaded, &decoded)
		Expect(err).NotTo(HaveOccurred())

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
		Ω(err).ShouldNot(HaveOccurred())
		Ω(results).Should(HaveLen(3))
		Ω(results[0].(*myResponse).Foo).Should(Equal("foo"))
	})

	It("Loads results between two particular dates", func() {
		save()
		b := save()
		c := save()
		d := save()
		save()

		results, err := history.LoadBetween("test-runs", reflect.TypeOf(myResponse{}), b.Timestamp, d.Timestamp)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(results).Should(HaveLen(3))
		Ω(results[0].(*myResponse).Timestamp.UnixNano()).Should(Equal(b.Timestamp.UnixNano()))
		Ω(results[1].(*myResponse).Timestamp.UnixNano()).Should(Equal(c.Timestamp.UnixNano()))
		Ω(results[2].(*myResponse).Timestamp.UnixNano()).Should(Equal(d.Timestamp.UnixNano()))
	})
})
