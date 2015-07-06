package lib

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type limitedReq struct {
	w    http.ResponseWriter
	req  *http.Request
	done chan struct{}
}

func limitHandler(num int, h http.Handler) http.Handler {
	if num <= 0 {
		return h
	}

	ch := make(chan *limitedReq, num)

	go func(ch chan *limitedReq) {
		for r := range ch {
			h.ServeHTTP(r.w, r.req)
			if r.done != nil {
				r.done <- struct{}{}
			}
		}
	}(ch)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		done := make(chan struct{})
		ch <- &limitedReq{w, req, done}
		<-done
	})
}

func registerDirectoryWithFallback(cfg Path, mux Mux, log LogFn) {
	log("Path: %v", cfg.Prefix)
	log("Directory: %v", cfg.Directory)
	log("Strip prefix? %v", cfg.StripPrefix)
	log("Fallback: %v", cfg.ReverseProxy)

	fallback := setupReverseProxy(cfg, log)

	// this handler is quite simple
	// basically this is a FileServe that combines the
	// actual directory on cfg and the files avialable using GAS
	server := func(w http.ResponseWriter, req *http.Request) {
		p := req.URL.Path
		// check if the file exists
		file := filepath.Join(cfg.Directory, filepath.FromSlash(p))
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
			log("Using fallback for %v", p)
			fallback.ServeHTTP(w, req)
			return
		}
		http.ServeFile(w, req, file)
		return
	}

	var handler http.Handler
	handler = http.HandlerFunc(server)

	if cfg.StripPrefix {
		handler = http.StripPrefix(cfg.Prefix, http.HandlerFunc(server))
	}
	mux.Handle(cfg.Prefix, logHandler(handler, log))
}

// Mux interface to register paths
type Mux interface {
	Handle(path string, hander http.Handler)
}

// Type used to log information
type LogFn func(format string, args ...interface{})

func registerDirectory(p Path, mux Mux, log LogFn) {
	log("Path: %v", p.Prefix)
	log("Directory: %v", p.Directory)
	log("Strip prefix? %v", p.StripPrefix)

	dir := http.Dir(p.Directory)
	var handler http.Handler
	handler = http.FileServer(dir)
	if p.StripPrefix {
		handler = http.StripPrefix(p.Prefix, handler)
	}
	mux.Handle(p.Prefix, logHandler(handler, log))
}

func registerReverseProxy(p Path, endpoint string, mux Mux, log LogFn) {
	handler := setupReverseProxy(p, log)
	mux.Handle(p.Prefix, logHandler(handler, log))
}

func setupReverseProxy(p Path, log LogFn) http.Handler {
	log("Path: %v", p.Prefix)
	log("Endpoint: %v", p.ReverseProxy)
	log("Strip prefix? %v", p.StripPrefix)

	revUrl, err := url.Parse(p.ReverseProxy)
	if err != nil {
		log("Error parsing endpoint url: %v", err)
	}

	var handler http.Handler
	handler = httputil.NewSingleHostReverseProxy(revUrl)
	if p.StripPrefix {
		handler = http.StripPrefix(p.Prefix, handler)
	}

	return limitHandler(p.Limit, handler)
}

// Register the given configuration on the provided mux
func RegisterConfig(mux Mux, cfg *Config, log LogFn) {
	for _, p := range cfg.Paths {
		if len(p.Directory) > 0 {
			if p.WithFallback {
				registerDirectoryWithFallback(p, mux, log)
			} else {
				registerDirectory(p, mux, log)
			}
		} else if len(p.ReverseProxy) > 0 {
			registerReverseProxy(p, p.ReverseProxy, mux, log)
		}
	}
}

// Serve the root path (/) from the given file path
func ServeRootFrom(m Mux, path string, log LogFn) {
	http.Handle("/", logHandler(http.FileServer(http.Dir(path)), log))
}

// wrapp the handler to print information on stdout
func logHandler(handler http.Handler, log LogFn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log("%v - %v", req.Method, req.URL)
		handler.ServeHTTP(w, req)
	})
}

// Wrap the given handler and remove the If-Modified-Since header before
// calling the hander
func DisableCacheHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Header.Del("If-Modified-Since")
		handler.ServeHTTP(w, req)
	})
}
