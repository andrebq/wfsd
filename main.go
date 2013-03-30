package main

import (
	"flag"
	"log"
	"net/http"
	"github.com/andrebq/wfsd/lib"
)

var (
	port = flag.String("p", ":8081", "Port to listen")
	cfg = flag.String("cfg", "wfsd.config", "Config file")
	disableCache = flag.Bool("disableCache", false, "Make wfsd ignore the If-Modified-Since Header")
)


func main() {
	flag.Parse()
	wfsdCfg, err := lib.Load(*cfg)
	if err != nil {
		log.Printf("Error loading config file. Cause: %v", err)
	}
	lib.RegisterConfig(http.DefaultServeMux, wfsdCfg, log.Printf)
	lib.ServeRootFrom(http.DefaultServeMux, ".", log.Printf)

	log.Printf("Starting server at %v", *port)
	if *disableCache {
		log.Printf("Cache is now disabled...")
		err = http.ListenAndServe(*port, createIgnoreCacheHandler(http.DefaultServeMux))
	} else {
		err = http.ListenAndServe(*port, nil)
	}

	if err != nil {
		log.Printf("Unable to start server. Cause: %v", err)
	}
}

// wrap handler with a new hander that will remove the If-Modified-Since header
func createIgnoreCacheHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Header.Del("If-Modified-Since")
		handler.ServeHTTP(w, req)
	})
}

