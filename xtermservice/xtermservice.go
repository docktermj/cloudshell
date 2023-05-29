package xtermservice

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"time"

	"github.com/docktermj/cloudshell/pkg/xtermjs"
	"github.com/gorilla/mux"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// XtermServiceImpl is the default implementation of the HttpServer interface.
type XtermServiceImpl struct {
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

func (xtermService *XtermServiceImpl) Handler(ctx context.Context) *http.ServeMux {

	// configure routing
	// router := mux.NewRouter()

	// this is the endpoint for xterm.js to connect to
	xtermjsHandlerOptions := xtermjs.HandlerOpts{
		AllowedHostnames:     xtermService.AllowedHostnames,
		Arguments:            xtermService.Arguments,
		Command:              xtermService.Command,
		ConnectionErrorLimit: xtermService.ConnectionErrorLimit,
		CreateLogger: func(connectionUUID string, r *http.Request) xtermjs.Logger {
			createRequestLog(r, map[string]interface{}{"connection_uuid": connectionUUID}).Infof("created logger for connection '%s'", connectionUUID)
			return createRequestLog(nil, map[string]interface{}{"connection_uuid": connectionUUID})
		},
		KeepalivePingTimeout: time.Duration(xtermService.KeepalivePingTimeout) * time.Second,
		MaxBufferSizeBytes:   xtermService.MaxBufferSizeBytes,
	}

	// fmt.Printf(">>>>>> xtermjs.GetHandler(xtermjsHandlerOptions): %s\n", reflect.TypeOf(xtermjs.GetHandler(xtermjsHandlerOptions)))

	submux := http.NewServeMux()
	submux.HandleFunc("/", xtermjs.GetHandler(xtermjsHandlerOptions))

	// router.HandleFunc(xtermService.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	// // this is the endpoint for serving xterm.js assets
	// depenenciesDirectory := path.Join(xtermService.WorkingDir, "./node_modules")
	// router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

	// // this is the endpoint for the root path aka website
	// publicAssetsDirectory := path.Join(xtermService.WorkingDir, "./public")
	// router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

	// // start memory logging pulse
	// logWithMemory := createMemoryLog()
	// go func(tick *time.Ticker) {
	// 	for {
	// 		logWithMemory.Debug("tick")
	// 		<-tick.C
	// 	}
	// }(time.NewTicker(time.Second * 30))

	return submux
}

/*
The Serve method simply prints the 'Something' value in the type-struct.

Input
  - ctx: A context to control lifecycle.

Output
  - Nothing is returned, except for an error.  However, something is printed.
    See the example output.
*/

func (xtermService *XtermServiceImpl) Serve(ctx context.Context) error {
	var err error = nil

	// configure routing
	router := mux.NewRouter()

	// this is the endpoint for xterm.js to connect to
	xtermjsHandlerOptions := xtermjs.HandlerOpts{
		AllowedHostnames:     xtermService.AllowedHostnames,
		Arguments:            xtermService.Arguments,
		Command:              xtermService.Command,
		ConnectionErrorLimit: xtermService.ConnectionErrorLimit,
		CreateLogger: func(connectionUUID string, r *http.Request) xtermjs.Logger {
			createRequestLog(r, map[string]interface{}{"connection_uuid": connectionUUID}).Infof("created logger for connection '%s'", connectionUUID)
			return createRequestLog(nil, map[string]interface{}{"connection_uuid": connectionUUID})
		},
		KeepalivePingTimeout: time.Duration(xtermService.KeepalivePingTimeout) * time.Second,
		MaxBufferSizeBytes:   xtermService.MaxBufferSizeBytes,
	}

	fmt.Printf(">>>>>> xtermjs.GetHandler(xtermjsHandlerOptions): %s\n", reflect.TypeOf(xtermjs.GetHandler(xtermjsHandlerOptions)))

	router.HandleFunc(xtermService.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	// this is the endpoint for serving xterm.js assets
	depenenciesDirectory := path.Join(xtermService.WorkingDir, "./node_modules")
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

	// this is the endpoint for the root path aka website
	publicAssetsDirectory := path.Join(xtermService.WorkingDir, "./public")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

	// start memory logging pulse
	logWithMemory := createMemoryLog()
	go func(tick *time.Ticker) {
		for {
			logWithMemory.Debug("tick")
			<-tick.C
		}
	}(time.NewTicker(time.Second * 30))

	return err
}
