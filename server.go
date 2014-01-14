package pat

import (
	"encoding/json"
	"fmt"
	"github.com/julz/pat/benchmarker"
	"github.com/julz/pat/history"
	"net/http"
	"reflect"
)

func Serve() {
	ServeWithArgs("historical-runs")
}

func ServeWithArgs(baseDir string) {
	ctx := &context{baseDir}
	http.HandleFunc("/experiments", handler(ctx, handleList))
	http.HandleFunc("/experiments/push", handler(ctx, handlePush))
}

func Stop() {
}

func Bind() {
	http.ListenAndServe(":8080", nil)
}

type listResponse struct {
	Items []interface{}
}

type context struct {
	baseDir string
}

func handleList(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	response, err := history.LoadAll(ctx.baseDir, reflect.TypeOf(Response{}))
	return &listResponse{response}, err
}

func handlePush(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	totalTime := benchmarker.Time(push)
	return history.Save(ctx.baseDir, &Response{totalTime.Nanoseconds()})
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
