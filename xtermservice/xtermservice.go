package xtermservice

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/docktermj/cloudshell/pkg/xtermjs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	ServerPort           int
	WorkingDir           string
}

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

//go:embed static/*
var static embed.FS

// ----------------------------------------------------------------------------
// Interface methods
// ----------------------------------------------------------------------------

/*
The Handler method...

Input
  - ctx: A context to control lifecycle.

Output
  - http.ServeMux...
*/

// func (xtermService *XtermServiceImpl) Handler(ctx context.Context) error {
func (xtermService *XtermServiceImpl) Handler(ctx context.Context) *http.ServeMux {
	rootMux := http.NewServeMux()

	// Add route to xterm.js.

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
	rootMux.HandleFunc(xtermService.PathXtermjs, xtermjs.GetHandler(xtermjsHandlerOptions))

	// Add route to readiness probe endpoint.

	rootMux.HandleFunc(xtermService.PathReadiness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Add route to liveness probe endpoint.

	rootMux.HandleFunc(xtermService.PathLiveness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Add route to metrics.

	rootMux.Handle(xtermService.PathMetrics, promhttp.Handler())

	// Add directory of static files.

	rootDir, err := fs.Sub(static, "static/root")
	if err != nil {
		panic(err)
	}
	rootMux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(rootDir))))

	return rootMux
}
