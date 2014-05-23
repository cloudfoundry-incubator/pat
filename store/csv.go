package store

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/cloudfoundry-incubator/pat/logs"
	"github.com/cloudfoundry-incubator/pat/workloads"
)

type CsvStore struct {
	dir          string
	workloadList *workloads.WorkloadList
}

type csvFile struct {
	outputPath string
	guid       string
	commands   []string
}

func NewCsvStore(dir string, list *workloads.WorkloadList) *CsvStore {
	return &CsvStore{dir, list}
}

func (store *CsvStore) Writer(guid string) func(samples <-chan *experiment.Sample) {
	return store.newCsvFile(guid).Write
}

func (store *CsvStore) load(filename string, guid string) (experiment.Experiment, error) {
	return &csvFile{path.Join(store.dir, filename), guid, nil}, nil
}

func (store *CsvStore) newCsvFile(guid string) *csvFile {
	file := &csvFile{path.Join(store.dir, strconv.Itoa(int(time.Now().UnixNano()))+"-"+guid+".csv"), guid, nil}
	store.workloadList.DescribeWorkloads(file)
	return file
}

func (file *csvFile) AddWorkloadStep(workload workloads.WorkloadStep) {
	file.commands = append(file.commands, workload.Name)
}

func (self *csvFile) Write(samples <-chan *experiment.Sample) {
	var logger = logs.NewLogger("store.csv")

	f, err := os.Create(self.outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Creating directory, %s", filepath.Dir(self.outputPath))
			os.MkdirAll(filepath.Dir(self.outputPath), 0755)
			f, err = os.Create(self.outputPath)
		}

		if err != nil {
			logger.Errorf("Can't write CSV: %v", err)
		}
	}
	defer f.Close()

	var header []string
	var body []string
	w := csv.NewWriter(f)

	header = []string{"Average", "TotalTime", "Total", "TotalErrors", "LastError", "TotalWorkers", "LastResult", "WorstResult", "NinetyfifthPercentile", "WallTime", "Type"}
	for _, k := range self.commands {
		header = append(header, "Commands|"+k+"|Count",
			"Commands|"+k+"|Throughput",
			"Commands|"+k+"|Average",
			"Commands|"+k+"|TotalTime",
			"Commands|"+k+"|LastTime",
			"Commands|"+k+"|WorstTime")
	}
	w.Write(header)

	for s := range samples {
		if s.Type == experiment.ResultSample {
			body = []string{strconv.Itoa(int(s.Average.Nanoseconds())),
				strconv.Itoa(int(s.TotalTime.Nanoseconds())),
				strconv.Itoa(int(s.Total)),
				strconv.Itoa(int(s.TotalErrors)),
				s.LastError,
				strconv.Itoa(int(s.TotalWorkers)),
				strconv.Itoa(int(s.LastResult.Nanoseconds())),
				strconv.Itoa(int(s.WorstResult.Nanoseconds())),
				strconv.Itoa(int(s.NinetyfifthPercentile.Nanoseconds())),
				strconv.Itoa(int(s.WallTime)),
				strconv.Itoa(int(s.Type))}

			for _, k := range self.commands {
				if s.Commands[k].Count == 0 {
					body = append(body, "", "", "", "", "", "")
				} else {
					body = append(body, strconv.Itoa(int(s.Commands[k].Count)),
						strconv.FormatFloat(s.Commands[k].Throughput, 'f', 8, 64),
						strconv.Itoa(int(s.Commands[k].Average.Nanoseconds())),
						strconv.Itoa(int(s.Commands[k].TotalTime.Nanoseconds())),
						strconv.Itoa(int(s.Commands[k].LastTime.Nanoseconds())),
						strconv.Itoa(int(s.Commands[k].WorstTime.Nanoseconds())))
				}
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
	var cmdColumns = make(map[string]int)
	for i, d := range decoded {
		if i == 0 {
			for n, s := range d {
				if strings.HasPrefix(s, "Commands|") {
					cmdColumns[s] = n
				}
			}
		} else {
			sample := &experiment.Sample{}
			sample.Commands = make(map[string]experiment.Command)
			sample.Average, err = duration(d[0])
			sample.TotalTime, err = duration(d[1])
			sample.Total, err = i64(d[2])
			sample.TotalErrors, err = strconv.Atoi(d[3])
			sample.LastError = d[4]
			sample.TotalWorkers, err = strconv.Atoi(d[5])
			sample.LastResult, err = duration(d[6])
			sample.WorstResult, err = duration(d[7])
			sample.NinetyfifthPercentile, err = duration(d[8])
			sample.WallTime, err = duration(d[9])
			sample.Type = experiment.ResultSample // this is the only type we currently persist

			var cmdName string
			for k, _ := range cmdColumns {
				if strings.Split(k, "|")[2] != "Count" {
					continue
				}
				cmdName = strings.Split(k, "|")[1]
				cmd.Count, err = i64(d[cmdColumns["Commands|"+cmdName+"|Count"]])
				if cmd.Count > 0 {
					cmd.Throughput, err = strconv.ParseFloat(d[cmdColumns["Commands|"+cmdName+"|Throughput"]], 64)
					cmd.Average, err = duration(d[cmdColumns["Commands|"+cmdName+"|Average"]])
					cmd.TotalTime, err = duration(d[cmdColumns["Commands|"+cmdName+"|TotalTime"]])
					cmd.LastTime, err = duration(d[cmdColumns["Commands|"+cmdName+"|LastTime"]])
					cmd.WorstTime, err = duration(d[cmdColumns["Commands|"+cmdName+"|WorstTime"]])
					sample.Commands[cmdName] = cmd
				} else {
					err = nil //reset the expected error for empty fields
				}
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
