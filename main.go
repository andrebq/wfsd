package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	port = flag.String("p", ":8081", "Port to listen")
	cfg = flag.String("cfg", "wfsd.config", "Config file")
)

func main() {
	flag.Parse()
	cfg, err := Load(*cfg)
	if err != nil {
		log.Printf("Error loading config file. Cause: %v", err)
	}

	for _, p := range cfg.Paths {
		log.Printf("Path: %v", p.Prefix)
		log.Printf("Directory: %v", p.Directory)
		log.Printf("Strip prefix? %v", p.StripPrefix)

		dir := http.Dir(p.Directory)
		handler := http.FileServer(dir)
		if p.StripPrefix {
			handler = http.StripPrefix(p.Prefix, handler)
		}
		http.Handle(p.Prefix, logHandler(handler))
	}

	http.Handle("/", logHandler(http.FileServer(http.Dir("."))))

	log.Printf("Starting server at %v", *port)
	err = http.ListenAndServe(*port, nil)
	if err != nil {
		log.Printf("Unable to start server. Cause: %v", err)
	}
}

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%v - %v", req.Method, req.URL)
		handler.ServeHTTP(w, req)
	})
}
