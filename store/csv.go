package store

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-community/pat/experiment"
)

type CsvStore struct {
	dir string
}

type csvFile struct {
	outputPath string
	guid       string
}

func NewCsvStore(dir string) *CsvStore {
	return &CsvStore{dir}
}

func (store *CsvStore) Writer(guid string) func(samples <-chan *experiment.Sample) {
	return store.newCsvFile(guid).Write
}

func (store *CsvStore) load(filename string, guid string) (experiment.Experiment, error) {
	return &csvFile{path.Join(store.dir, filename), guid}, nil
}

func (store *CsvStore) newCsvFile(guid string) *csvFile {
	return &csvFile{path.Join(store.dir, strconv.Itoa(int(time.Now().UnixNano()))+"-"+guid+".csv"), guid}
}

func (self *csvFile) Write(samples <-chan *experiment.Sample) {
	f, err := os.Create(self.outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating directory, ", filepath.Dir(self.outputPath))
			os.MkdirAll(filepath.Dir(self.outputPath), 0755)
			f, err = os.Create(self.outputPath)
		}

		if err != nil {
			fmt.Println("Can't write CSV: ", err)
		}
	}
	defer f.Close()

	var header []string
	var body []string
	w := csv.NewWriter(f)

	for s := range samples {
		if s.Type == experiment.ResultSample {

			if len(header) == 0 {
				header = []string{"Average", "TotalTime", "Total", "TotalErrors", "TotalWorkers", "LastResult", "WorstResult", "NinetyfifthPercentile", "WallTime", "Type"}
				for k, _ := range s.Commands {
					header = append(header, "Commands:"+k+":Count",
						"Commands:"+k+":Throughput",
						"Commands:"+k+":Average",
						"Commands:"+k+":TotalTime",
						"Commands:"+k+":LastTime",
						"Commands:"+k+":WorstTime")
				}
				w.Write(header)
			}

			body = []string{strconv.Itoa(int(s.Average.Nanoseconds())),
				strconv.Itoa(int(s.TotalTime.Nanoseconds())),
				strconv.Itoa(int(s.Total)),
				strconv.Itoa(int(s.TotalErrors)),
				strconv.Itoa(int(s.TotalWorkers)),
				strconv.Itoa(int(s.LastResult.Nanoseconds())),
				strconv.Itoa(int(s.WorstResult.Nanoseconds())),
				strconv.Itoa(int(s.NinetyfifthPercentile.Nanoseconds())),
				strconv.Itoa(int(s.WallTime)),
				strconv.Itoa(int(s.Type))}

			for k, _ := range s.Commands {
				body = append(body, strconv.Itoa(int(s.Commands[k].Count)),
					strconv.FormatFloat(s.Commands[k].Throughput, 'f', 8, 64),
					strconv.Itoa(int(s.Commands[k].Average.Nanoseconds())),
					strconv.Itoa(int(s.Commands[k].TotalTime.Nanoseconds())),
					strconv.Itoa(int(s.Commands[k].LastTime.Nanoseconds())),
					strconv.Itoa(int(s.Commands[k].WorstTime.Nanoseconds())))
			}

			w.Write(body)
			w.Flush()
		}
	}
}

func (self *csvFile) GetData() (samples []*experiment.Sample, err error) {
	file, err := os.Open(self.outputPath)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	decoded, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	var cmd experiment.Command
	var keys = make(map[string]bool)
	for i, d := range decoded {
		if i == 0 {
			for _, s := range d {
				if strings.HasPrefix(s, "Commands:") {
					keys[strings.Split(s, ":")[1]] = true
				}
			}
		} else {
			sample := &experiment.Sample{}
			sample.Commands = make(map[string]experiment.Command)
			sample.Average, err = duration(d[0])
			sample.TotalTime, err = duration(d[1])
			sample.Total, err = i64(d[2])
			sample.TotalErrors, err = strconv.Atoi(d[3])
			sample.TotalWorkers, err = strconv.Atoi(d[4])
			sample.LastResult, err = duration(d[5])
			sample.WorstResult, err = duration(d[6])
			sample.NinetyfifthPercentile, err = duration(d[7])
			sample.WallTime, err = duration(d[8])
			sample.Type = experiment.ResultSample // this is the only type we currently persist

			var i = 10
			for k, _ := range keys {
				cmd.Count, err = i64(d[i])
				cmd.Throughput, err = strconv.ParseFloat(d[i+1], 64)
				cmd.Average, err = duration(d[i+2])
				cmd.TotalTime, err = duration(d[i+3])
				cmd.LastTime, err = duration(d[i+4])
				cmd.WorstTime, err = duration(d[i+5])
				sample.Commands[k] = cmd
				i += 6
			}

			if err != nil {
				return nil, err
			}

			samples = append(samples, sample)
		}
	}
	return
}

func (store *CsvStore) LoadAll() (samples []experiment.Experiment, err error) {
	files, err := ioutil.ReadDir(store.dir)
	if err != nil {
		return nil, err
	}

	samples = make([]experiment.Experiment, 0)
	for _, f := range files {
		base := strings.Split(f.Name(), ".")[0]
		name := strings.SplitN(base, "-", 2)[1]
		if len(name) > 0 {
			loaded, err := store.load(f.Name(), name)
			if err == nil {
				samples = append(samples, loaded)
			}
		}
	}

	return
}

func (csv *csvFile) GetGuid() string {
	return csv.guid
}

func i64(s string) (int64, error) {
	t, e := strconv.Atoi(s)
	return int64(t), e
}

func duration(s string) (time.Duration, error) {
	t, e := strconv.Atoi(s)
	return time.Duration(t) * time.Nanosecond, e
}
