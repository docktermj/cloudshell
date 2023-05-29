package xtermserver

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"time"

	"github.com/docktermj/cloudshell/pkg/xtermjs"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// XtermServerImpl is the default implementation of the HttpServer interface.
type XtermServerImpl struct {
	AllowedHostnames     []string
	Arguments            []string
	Command              string
	ConnectionErrorLimit int
	KeepalivePingTimeout int
	MaxBufferSizeBytes   int
	PathLiveness         string
	PathMetrics          string
	PathReadiness        string
	PathXtermjs          string
	ServerAddr           string
	Port                 int
	WorkingDir           string
}

// ----------------------------------------------------------------------------
// Interface methods
// ----------------------------------------------------------------------------

/*
The Serve method simply prints the 'Something' value in the type-struct.

Input
  - ctx: A context to control lifecycle.

Output
  - Nothing is returned, except for an error.  However, something is printed.
    See the example output.
*/

func (httpServer *XtermServerImpl) Serve(ctx context.Context) error {
	rootMux := http.NewServeMux()

	// configure routing
	// router := mux.NewRouter()

	// fmt.Printf(">>>>>> router: %s\n", reflect.TypeOf(router))

	// this is the endpoint for xterm.js to connect to
	xtermjsHandlerOptions := xtermjs.HandlerOpts{
		AllowedHostnames:     httpServer.AllowedHostnames,
		Arguments:            httpServer.Arguments,
		Command:              httpServer.Command,
		ConnectionErrorLimit: httpServer.ConnectionErrorLimit,
		CreateLogger: func(connectionUUID string, r *http.Request) xtermjs.Logger {
			createRequestLog(r, map[string]interface{}{"connection_uuid": connectionUUID}).Infof("created logger for connection '%s'", connectionUUID)
			return createRequestLog(nil, map[string]interface{}{"connection_uuid": connectionUUID})
		},
		KeepalivePingTimeout: time.Duration(httpServer.KeepalivePingTimeout) * time.Second,
		MaxBufferSizeBytes:   httpServer.MaxBufferSizeBytes,
	}

	// fmt.Printf(">>>>>> xtermjsHandlerOptions: %s\n", reflect.TypeOf(xtermjsHandlerOptions))

	rootMux.HandleFunc(httpServer.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	// fmt.Printf(">>>>>> xtermjs.GetHandler(xtermjsHandlerOptions): %s\n", xtermjs.GetHandler(xtermjsHandlerOptions))

	// metrics endpoint
	rootMux.Handle(httpServer.PathMetrics, promhttp.Handler())

	// this is the endpoint for serving xterm.js assets
	depenenciesDirectory := path.Join(httpServer.WorkingDir, "./node_modules")
	rootMux.Handle("/assets", http.FileServer(http.Dir(depenenciesDirectory)))
	// rootMux.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

	// this is the endpoint for the root path aka website
	publicAssetsDirectory := path.Join(httpServer.WorkingDir, "./public")
	rootMux.Handle("/", http.FileServer(http.Dir(publicAssetsDirectory)))
	// rootMux.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

	// Start service.

	// if err := http.ListenAndServe(fmt.Sprintf(":%d", httpServer.Port), rootMux); err != nil {
	// 	log.Fatal(err)
	// }
	// return err

	// listen

	listenOnAddress := fmt.Sprintf("%s:%v", httpServer.ServerAddr, httpServer.Port)
	server := http.Server{
		Addr:    listenOnAddress,
		Handler: addIncomingRequestLogging(rootMux),
	}

	fmt.Printf("starting server on interface:port '%s'...", listenOnAddress)
	return server.ListenAndServe()
}

/*
The Serve method simply prints the 'Something' value in the type-struct.

Input
  - ctx: A context to control lifecycle.

Output
  - Nothing is returned, except for an error.  However, something is printed.
    See the example output.
*/

func (httpServer *XtermServerImpl) ServeOriginal(ctx context.Context) error {

	// configure routing
	router := mux.NewRouter()

	fmt.Printf(">>>>>> router: %s\n", reflect.TypeOf(router))

	// this is the endpoint for xterm.js to connect to
	xtermjsHandlerOptions := xtermjs.HandlerOpts{
		AllowedHostnames:     httpServer.AllowedHostnames,
		Arguments:            httpServer.Arguments,
		Command:              httpServer.Command,
		ConnectionErrorLimit: httpServer.ConnectionErrorLimit,
		CreateLogger: func(connectionUUID string, r *http.Request) xtermjs.Logger {
			createRequestLog(r, map[string]interface{}{"connection_uuid": connectionUUID}).Infof("created logger for connection '%s'", connectionUUID)
			return createRequestLog(nil, map[string]interface{}{"connection_uuid": connectionUUID})
		},
		KeepalivePingTimeout: time.Duration(httpServer.KeepalivePingTimeout) * time.Second,
		MaxBufferSizeBytes:   httpServer.MaxBufferSizeBytes,
	}

	fmt.Printf(">>>>>> xtermjsHandlerOptions: %s\n", reflect.TypeOf(xtermjsHandlerOptions))

	router.HandleFunc(httpServer.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	fmt.Printf(">>>>>> xtermjs.GetHandler(xtermjsHandlerOptions): %s\n", xtermjs.GetHandler(xtermjsHandlerOptions))

	// readiness probe endpoint
	router.HandleFunc(httpServer.PathReadiness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// liveness probe endpoint
	router.HandleFunc(httpServer.PathLiveness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// metrics endpoint
	router.Handle(httpServer.PathMetrics, promhttp.Handler())

	// this is the endpoint for serving xterm.js assets
	depenenciesDirectory := path.Join(httpServer.WorkingDir, "./node_modules")
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

	// this is the endpoint for the root path aka website
	publicAssetsDirectory := path.Join(httpServer.WorkingDir, "./public")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

	// start memory logging pulse
	logWithMemory := createMemoryLog()
	go func(tick *time.Ticker) {
		for {
			logWithMemory.Debug("tick")
			<-tick.C
		}
	}(time.NewTicker(time.Second * 30))

	// listen
	listenOnAddress := fmt.Sprintf("%s:%v", httpServer.ServerAddr, httpServer.Port)
	server := http.Server{
		Addr:    listenOnAddress,
		Handler: addIncomingRequestLogging(router),
	}

	fmt.Printf("starting server on interface:port '%s'...", listenOnAddress)
	return server.ListenAndServe()
}
