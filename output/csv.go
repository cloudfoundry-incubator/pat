package output

import (
  "encoding/csv"
  "fmt"
  "github.com/julz/pat/experiment"
  "os"
  "path/filepath"
  "strconv"
)

type CsvSampleWriter struct {
  output string
}

type Writer interface {
  Write(samples chan *experiment.Sample)
}

func NewCsvWriter(name string) Writer {
  return &CsvSampleWriter{name}
}

func (self *CsvSampleWriter) Write(samples chan *experiment.Sample) {
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
  w.Write([]string{"duration", "wallTime", "average", "workers"})

  for s := range samples {
    if s.Type == experiment.ResultSample {
      w.Write([]string{strconv.Itoa(int(s.LastResult.Nanoseconds())), strconv.Itoa(int(s.WallTime)), strconv.Itoa(int(s.Average.Nanoseconds())), strconv.Itoa(int(s.TotalWorkers))})
      w.Flush()
    }
  }
}
