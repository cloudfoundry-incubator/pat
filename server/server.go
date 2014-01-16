package server

import (
	"encoding/json"
	"fmt"
	"github.com/julz/pat/benchmarker"
	"github.com/julz/pat/experiments"
	"github.com/julz/pat/history"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func Serve() {
	ServeWithArgs("historical-runs")
}

func ServeWithArgs(baseDir string) {
	ctx := &context{baseDir}
	http.HandleFunc("/experiments/", handler(ctx, handleList))
	http.HandleFunc("/experiments/push", handler(ctx, handlePush))
}

func Stop() {
}

func Bind() {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("ui"))))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("ListenAndServe: %s\n", err)
	}
	fmt.Println("Started listening on :8080")
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
	totalTime, _ := benchmarker.Time(experiments.Dummy)
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
