package output

import (
	"encoding/csv"
	"fmt"
	"github.com/julz/pat/experiment"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CsvSampleFile struct {
	output string
}

func NewCsvWriter(name string) *CsvSampleFile {
	return &CsvSampleFile{name}
}

func (self *CsvSampleFile) Write(samples chan *experiment.Sample) {
	f, err := os.Create(self.output)
	defer f.Close()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating directory, ", self.output)
			os.Mkdir(filepath.Dir(self.output), 0755)
			f, err = os.Create(self.output)
		}

		if err != nil {
			fmt.Println("Can't write CSV: ", err)
		}
	}

	w := csv.NewWriter(f)
	w.Write([]string{"Average", "TotalTime", "Total", "TotalErrors", "TotalWorkers", "LastResult", "WorstResult", "WallTime", "Type"})

	for s := range samples {
		if s.Type == experiment.ResultSample {
			w.Write([]string{strconv.Itoa(int(s.Average.Nanoseconds())),
				strconv.Itoa(int(s.TotalTime.Nanoseconds())),
				strconv.Itoa(int(s.Total)),
				strconv.Itoa(int(s.TotalErrors)),
				strconv.Itoa(int(s.TotalWorkers)),
				strconv.Itoa(int(s.LastResult.Nanoseconds())),
				strconv.Itoa(int(s.WorstResult.Nanoseconds())),
				strconv.Itoa(int(s.WallTime)),
				strconv.Itoa(int(s.Type))})
			w.Flush()
		}
	}
}

func (self *CsvSampleFile) Read() (samples []*experiment.Sample, err error) {
	file, err := os.Open(self.output)
	defer file.Close()

	decoded, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	for i, d := range decoded {
		if i == 0 {
		} else {
			sample := &experiment.Sample{}
			sample.Average, err = duration(d[0])
			sample.TotalTime, err = duration(d[1])
			sample.Total, err = i64(d[2])
			sample.TotalErrors, err = strconv.Atoi(d[3])
			sample.TotalWorkers, err = strconv.Atoi(d[4])
			sample.LastResult, err = duration(d[5])
			sample.WorstResult, err = duration(d[6])
			sample.WallTime, err = duration(d[7])
			sample.Type = experiment.ResultSample // this is the only type we currently persist

			if err != nil {
				return nil, err
			}

			samples = append(samples, sample)
		}
	}
	return
}

func ReloadCSVs(baseDir string) (samples map[string][]*experiment.Sample, order []string, err error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, nil, err
	}

	samples = make(map[string][]*experiment.Sample)
	for _, f := range files {
		name := strings.Split(f.Name(), ".")[0]
		samples[name], err = NewCsvWriter(path.Join(baseDir, f.Name())).Read()
		order = append(order, name)
	}

	return
}

func i64(s string) (int64, error) {
	t, e := strconv.Atoi(s)
	return int64(t), e
}

func duration(s string) (time.Duration, error) {
	t, e := strconv.Atoi(s)
	return time.Duration(t) * time.Nanosecond, e
}
