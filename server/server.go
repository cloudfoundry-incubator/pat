package server

import (
  "encoding/json"
  "fmt"
  "github.com/gorilla/mux"
  "github.com/julz/pat/experiment"
  "github.com/julz/pat/history"
  "github.com/julz/pat/output"
  "github.com/nu7hatch/gouuid"
  "net/http"
  "net/url"
  "path"
  "reflect"
  "strconv"
  "time"
)

type Response struct {
  TotalTime int64
  Timestamp int64
}

type context struct {
  router  *mux.Router
  baseDir string
  csvDir  string
  running map[string][]*experiment.Sample
}

func Serve() {
  ServeWithArgs("historical-runs", "output/csvs")
}

func ServeWithArgs(baseDir string, csvDir string) {
  r := mux.NewRouter()
  ctx := &context{r, baseDir, csvDir, make(map[string][]*experiment.Sample)}
  err := ctx.reload()
  if err != nil {
    fmt.Println("Couldn't load previous experiments, ", err)
  }

  r.Methods("GET").Path("/experiments/").HandlerFunc(handler(ctx.handleListExperiments))
  r.Methods("GET").Path("/experiments/{name}").HandlerFunc(handler(ctx.handleGetExperiment)).Name("experiment")
  r.Methods("POST").Path("/experiments/").HandlerFunc(handler(ctx.handlePush))

  http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
  http.Handle("/csv/experiments/", http.StripPrefix("/csv/experiments/", http.FileServer(http.Dir(ctx.csvDir))))
  http.Handle("/", r)
}

func Stop() {
}

func Bind() {
  fmt.Println("Starting web ui on http://localhost:8080/ui/")
  if err := http.ListenAndServe(":8080", nil); err != nil {
    fmt.Printf("ListenAndServe: %s\n", err)
  }
}

type listResponse struct {
  Items interface{}
}

func (ctx *context) reload() (err error) {
  // this is super-simple right now, we just load all the CSVs back in to memory
  // will move to using REDIS / SQLite at some point
  // also I'm aware server.go isn't well covered by tests and needs back-filling now that we
  // lost the previous system tests
  ctx.running, err = output.ReloadCSVs(ctx.csvDir)
  return err
}

func (ctx *context) handleList(w http.ResponseWriter, r *http.Request) (interface{}, error) {
  from, err := strconv.Atoi(r.FormValue("from"))
  to, err := strconv.Atoi(r.FormValue("to"))
  if err == nil {
    response, err := history.LoadBetween(ctx.baseDir, reflect.TypeOf(Response{}), time.Unix(0, int64(from)), time.Unix(0, int64(to)))
    return &listResponse{response}, err
  } else {
    response, err := history.LoadAll(ctx.baseDir, reflect.TypeOf(Response{}))
    return &listResponse{response}, err
  }
}

func (ctx *context) handleListExperiments(w http.ResponseWriter, r *http.Request) (interface{}, error) {
  running := make([]map[string]string, 0, len(ctx.running))
  for k, _ := range ctx.running {
    url, _ := ctx.router.Get("experiment").URL("name", k)
    csvUrl := fmt.Sprintf("/csv/%v.csv", url.String())
    json := make(map[string]string)
    json["Location"] = url.String()
    json["CsvLocation"] = csvUrl
    json["Name"] = "Simple Push"
    json["State"] = "Unknown"
    running = append(running, json)
  }

  return &listResponse{running}, nil
}

func (ctx *context) handlePush(w http.ResponseWriter, r *http.Request) (interface{}, error) {
  name, _ := uuid.NewV4()

  pushes, err := strconv.Atoi(r.FormValue("pushes"))
  if err != nil {
    pushes = 1
  }

  concurrency, err := strconv.Atoi(r.FormValue("concurrency"))
  if err != nil {
    concurrency = 1
  }

  handlers := make([]func(chan *experiment.Sample), 0)
  handlers = append(handlers, output.NewCsvWriter(path.Join(ctx.csvDir, name.String())+".csv").Write)
  handlers = append(handlers, func(samples chan *experiment.Sample) {
    ctx.buffer(name.String(), samples)
  })

  //ToDo (simon): interval and stop is 0, repeating at interval is not yet exposed in Web UI
  go experiment.Run(pushes, concurrency, 0, 0, output.Multiplexer(handlers).Multiplex)

  return ctx.router.Get("experiment").URL("name", name.String())
}

func (ctx *context) handleGetExperiment(w http.ResponseWriter, r *http.Request) (interface{}, error) {
  name := mux.Vars(r)["name"]
  // TODO(jz) only send back since N
  return &listResponse{ctx.running[name]}, nil
}

func (ctx *context) buffer(name string, samples chan *experiment.Sample) {
  for s := range samples {
    // FIXME(jz) - need to clear this at some point, memory leak..
    ctx.running[name] = append(ctx.running[name], s)
  }
}

func handler(fn func(w http.ResponseWriter, r *http.Request) (interface{}, error)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var err error
    var response interface{}
    var encoded []byte

    if response, err = fn(w, r); err == nil {
      switch r := response.(type) {
      case *url.URL:
        w.Header().Set("Location", r.String())
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "{ \"Location\": \"%v\", \"CsvLocation\": \"/csv/%v.csv\" }", r, r)
        return
      default:
        if encoded, err = json.Marshal(r); err == nil {
          w.Header().Set("Content-Type", "application/json")
          fmt.Fprintf(w, string(encoded))
          return
        }
      }
    }

    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}
