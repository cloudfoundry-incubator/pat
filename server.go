package pat

import (
  "encoding/json"
  "fmt"
  "github.com/julz/pat/benchmarker"
  "github.com/julz/pat/history"
  "net/http"
  "os"
  "reflect"
  "strconv"
  "time"
)

func Serve() {
  ServeWithArgs("historical-runs")
}

func ServeWithArgs(baseDir string) {
  ctx := &context{baseDir}
  http.HandleFunc("/experiments/", handler(ctx, handleList))
  http.HandleFunc("/experiments/push", handler(ctx, handlePush))
}

func serverWeb() {
  http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
  http.HandleFunc("/push", pushHandler)
}

func Stop() {
}

func Bind() {
  serverWeb()
  if err := http.ListenAndServe(":8080", nil); err != nil {
    fmt.Printf("ListenAndServe: %s\n", err)
  }
  fmt.Println("Started listening on :8080")
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
  //fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
  fmt.Fprintf(os.Stdout, "Hi there, I love %s!", r.URL.Path[1:])
  CallPush(w, 1, 1)
}

type listResponse struct {
  Items []interface{}
}

type context struct {
  baseDir string
}

func handleList(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
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

func handlePush(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
  totalTime, _ := benchmarker.Time(dummy)
  result := &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
  return history.Save(ctx.baseDir, result, result.Timestamp)
}

func handler(ctx *context, fn func(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var err error
    var response interface{}
    var encoded []byte
    if response, err = fn(ctx, w, r); err == nil {
      if encoded, err = json.Marshal(response); err == nil {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, string(encoded))
        return
      }
    }

    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}
