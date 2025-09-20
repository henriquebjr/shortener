package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/henriquebjr/shortener/url"
)

var (
	port    *int
	logOn   *bool
	baseUrl string
)

type Headers map[string]string

type Redirector struct {
	stats chan string
}

func init() {
	port = flag.Int("p,", 8888, "port")
	logOn = flag.Bool("l", true, "log on/off")

	flag.Parse()

	baseUrl = fmt.Sprintf("http://localhost:%d", *port)
}

func main() {
	stats := make(chan string)
	defer close(stats)
	go statsRegister(stats)

	url.Configure(url.NewRepository())

	http.HandleFunc("/api/shorten", Shortener)
	http.HandleFunc("/api/stats/", Viewer)
	http.Handle("/r/", &Redirector{stats})

	logger("Starting server in port %d...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func Shortener(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		answerWith(w, http.StatusMethodNotAllowed, Headers{
			"Allow": "POST",
		})
		return
	}

	url, isNew, err := url.FindOrCreateNewUrl(extractUrl(r))

	if err != nil {
		answerWith(w, http.StatusBadRequest, nil)
		return
	}

	shortUrl := fmt.Sprintf("%s/r/%s", baseUrl, url.Id)

	var status int
	if isNew {
		status = http.StatusCreated
		logger("Short URL %s succefully created to %s.", shortUrl, url.Destiny)
	} else {
		status = http.StatusOK
	}

	answerWith(w, status, Headers{
		"Location": shortUrl,
		"Link":     fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", baseUrl, url.Id),
	})
}

func answerWith(w http.ResponseWriter, status int, headers Headers) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
}

func extractUrl(r *http.Request) string {
	url := make([]byte, r.ContentLength, r.ContentLength)
	r.Body.Read(url)
	return string(url)
}

func (red *Redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	findUrlAndExecute(w, r, func(url *url.Url) {
		http.Redirect(w, r, url.Destiny, http.StatusMovedPermanently)

		red.stats <- url.Id
	})
}

func findUrlAndExecute(w http.ResponseWriter, r *http.Request, executor func(*url.Url)) {
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]

	if url := url.Find(id); url != nil {
		executor(url)
	} else {
		http.NotFound(w, r)
	}
}

func Viewer(w http.ResponseWriter, r *http.Request) {
	findUrlAndExecute(w, r, func(url *url.Url) {
		json, err := json.Marshal(url.Stats())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		answerWithJson(w, string(json))
	})
}

func answerWithJson(w http.ResponseWriter, answer string) {
	answerWith(w, http.StatusOK, Headers{"Content-Type": "application/json"})
	fmt.Fprint(w, answer)
}

func statsRegister(ids <-chan string) {
	for id := range ids {
		url.Register(id)
		logger("Redirect registered to %s.", id)
	}
}

func logger(format string, values ...interface{}) {
	if *logOn {
		log.Printf(fmt.Sprintf("%s\n", format), values...)
	}
}
