package history

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "os"
  "path"
  "reflect"
  "time"
)

func LoadAll(baseDir string, as reflect.Type) (all []interface{}, err error) {
  files, err := ioutil.ReadDir(baseDir)
  if err != nil {
    return nil, err
  }

  for _, f := range files {
    loaded, err := ioutil.ReadFile(path.Join(baseDir, f.Name()))
    var decoded = reflect.New(as).Interface()
    err = json.Unmarshal(loaded, decoded)
    if err != nil {
      return nil, err
    }

    all = append(all, decoded)
  }

  return all, nil
}

func Save(baseDir string, reply interface{}) (chain interface{}, err error) {
  encoded, err := json.Marshal(reply)
  if err != nil {
    return nil, err
  }

  if err = createIfNeeded(baseDir); err != nil {
    return nil, err
  }

  err = ioutil.WriteFile(path.Join(baseDir, filenameFromTimestamp()), encoded, 0644)
  return reply, err
}

func filenameFromTimestamp() string {
  return fmt.Sprintf("%d.json", time.Now().UnixNano())
}

func createIfNeeded(dir string) error {
  _, err := os.Stat(dir)
  if err != nil && os.IsNotExist(err) {
    fmt.Println("Creating directory, ", dir)
    return os.Mkdir(dir, 0755)
  }
  return nil
}
