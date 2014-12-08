package lib

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

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

func registerTCPProxy(p Path, mux Mux, log LogFn) {
	log("Ws Proxy at: %v", p.Prefix)
	mux.Handle(p.Prefix, logHandler(NewWSProxy(), log))
}

func registerReverseProxy(p Path, endpoint string, mux Mux, log LogFn) {
	log("Path: %v", p.Prefix)
	log("Endpoint: %v", endpoint)
	log("Strip prefix? %v", p.StripPrefix)

	revUrl, err := url.Parse(endpoint)
	if err != nil {
		log("Error parsing endpoint url: %v", err)
	}

	var handler http.Handler
	handler = httputil.NewSingleHostReverseProxy(revUrl)
	if p.StripPrefix {
		handler = http.StripPrefix(p.Prefix, handler)
	}

	mux.Handle(p.Prefix, logHandler(handler, log))
}

// Register the given configuration on the provided mux
func RegisterConfig(mux Mux, cfg *Config, log LogFn) {
	for _, p := range cfg.Paths {
		if len(p.Directory) > 0 {
			registerDirectory(p, mux, log)
		} else if p.TCPProxy {
			registerTCPProxy(p, mux, log)
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
