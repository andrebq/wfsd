package lib

import (
	"net/http"
)

// Mux interface to register paths
type Mux interface {
	Handle(path string, hander http.Handler)
}

// Type used to log information
type LogFn func(format string, args ...interface{})

// Register the given configuration on the provided mux
func RegisterConfig(mux Mux, cfg *Config, log LogFn) {
	for _, p := range cfg.Paths {
		log("Path: %v", p.Prefix)
		log("Directory: %v", p.Directory)
		log("Strip prefix? %v", p.StripPrefix)

		dir := http.Dir(p.Directory)
		handler := http.FileServer(dir)
		if p.StripPrefix {
			handler = http.StripPrefix(p.Prefix, handler)
		}
		mux.Handle(p.Prefix, logHandler(handler, log))
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
