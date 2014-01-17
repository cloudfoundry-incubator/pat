package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/julz/pat/experiment"
	"github.com/julz/pat/history"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"net/url"
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
	running map[string][]*experiment.Sample
}

func Serve() {
	ServeWithArgs("historical-runs")
}

func ServeWithArgs(baseDir string) {
	r := mux.NewRouter()
	ctx := &context{r, baseDir, make(map[string][]*experiment.Sample)}
	r.Methods("GET").Path("/experiments/").HandlerFunc(handler(ctx, handleList))
	r.Methods("GET").Path("/experiments/{name}").HandlerFunc(handler(ctx, handleGetExperiment)).Name("experiment")
	r.Methods("POST").Path("/experiments/").HandlerFunc(handler(ctx, handlePush))

	// BUG(jz) For easy web-browser testing, remove
	r.HandleFunc("/POST/experiments/", handler(ctx, handlePush))

	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
	http.Handle("/", r)
}

func Stop() {
}

func Bind() {
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("ListenAndServe: %s\n", err)
	}
	fmt.Println("Started listening on :8080")
}

type listResponse struct {
	Items interface{}
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
	fmt.Println("handlepush")
	name, _ := uuid.NewV4()

	pushes, err := strconv.Atoi(r.FormValue("pushes"))
	if err != nil {
		pushes = 1
	}

	concurrency, err := strconv.Atoi(r.FormValue("concurrency"))
	if err != nil {
		concurrency = 1
	}

	go experiment.Run(pushes, concurrency, func(samples chan *experiment.Sample, target int) {
		ctx.buffer(name.String(), samples, target)
	})

	return ctx.router.Get("experiment").URL("name", name.String())
}

func handleGetExperiment(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	name := mux.Vars(r)["name"]
	// TODO(jz) only send back since N
	return &listResponse{ctx.running[name]}, nil
}

func (context *context) buffer(name string, samples chan *experiment.Sample, target int) {
	for s := range samples {
		fmt.Println("Got sample, ", s)
		// FIXME(jz) - need to clear this at some point, memory leak..
		context.running[name] = append(context.running[name], s)
	}
}

func handler(ctx *context, fn func(ctx *context, w http.ResponseWriter, r *http.Request) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response interface{}
		var encoded []byte

		if response, err = fn(ctx, w, r); err == nil {
			fmt.Println("Response: ", response)
			switch r := response.(type) {
			case *url.URL:
				w.Header().Set("Location", r.String())
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, "{ \"Location\": \"%v\" }", r)
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
