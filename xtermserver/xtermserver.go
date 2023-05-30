package xtermserver

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"time"

	"github.com/docktermj/cloudshell/pkg/xtermjs"
	"github.com/docktermj/cloudshell/xtermservice"
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
The Serve method serves the XtermService over HTTP.

Input
  - ctx: A context to control lifecycle.
*/

func (httpServer *XtermServerImpl) Serve(ctx context.Context) error {
	rootMux := http.NewServeMux()

	// Add XtermService.

	xtermService := &xtermservice.XtermServiceImpl{
		AllowedHostnames:     httpServer.AllowedHostnames,
		Arguments:            httpServer.Arguments,
		Command:              httpServer.Command,
		ConnectionErrorLimit: httpServer.ConnectionErrorLimit,
		KeepalivePingTimeout: httpServer.KeepalivePingTimeout,
		MaxBufferSizeBytes:   httpServer.MaxBufferSizeBytes,
		PathLiveness:         httpServer.PathLiveness,
		PathMetrics:          httpServer.PathMetrics,
		PathReadiness:        httpServer.PathReadiness,
		PathXtermjs:          httpServer.PathXtermjs,
		ServerAddress:        httpServer.ServerAddr,
		ServerPort:           httpServer.Port,
		WorkingDir:           httpServer.WorkingDir,
	}
	xtermMux := xtermService.Handler(ctx)
	rootMux.Handle("/", xtermMux)

	// Start service.

	listenOnAddress := fmt.Sprintf("%s:%v", httpServer.ServerAddr, httpServer.Port)
	server := http.Server{
		Addr:    listenOnAddress,
		Handler: addIncomingRequestLogging(rootMux),
	}
	fmt.Printf("starting server on interface:port '%s'...", listenOnAddress)
	return server.ListenAndServe()
}

func (httpServer *XtermServerImpl) DeprecatedServeVersion2(ctx context.Context) error {
	rootMux := http.NewServeMux()

	// Add route to xterm.js.

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
	rootMux.HandleFunc(httpServer.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	// Add route to metrics.

	rootMux.Handle(httpServer.PathMetrics, promhttp.Handler())

	// Add route to xterm.js assets.

	depenenciesDirectory := path.Join(httpServer.WorkingDir, "./node_modules")
	rootMux.Handle("/assets", http.FileServer(http.Dir(depenenciesDirectory)))

	// Add route to website.

	publicAssetsDirectory := path.Join(httpServer.WorkingDir, "./public")
	rootMux.Handle("/", http.FileServer(http.Dir(publicAssetsDirectory)))

	// Start service.

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

func (httpServer *XtermServerImpl) DeprecatedServeOriginal(ctx context.Context) error {

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
