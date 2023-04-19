package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type config struct {
	port int
	env  string
}

func main() {
	fmt.Println("Starting API service")

	var cfg config
	flag.IntVar(&cfg.port, "port", 80, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("REQUEST:\n%s", string(reqDump))
		w.Write([]byte(""))
	})

	fmt.Println("Running webserver")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil))
}
