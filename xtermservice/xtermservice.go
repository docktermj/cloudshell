package xtermservice

import (
	"context"
	"net/http"
	"path"
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
	Port                 int
	WorkingDir           string
}

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

	// Add route to metrics.

	rootMux.Handle(xtermService.PathMetrics, promhttp.Handler())

	// Add route to xterm.js assets.

	depenenciesDirectory := path.Join(xtermService.WorkingDir, "./node_modules")
	rootMux.Handle("/assets", http.FileServer(http.Dir(depenenciesDirectory)))

	// Add route to website.

	publicAssetsDirectory := path.Join(xtermService.WorkingDir, "./public")
	rootMux.Handle("/", http.FileServer(http.Dir(publicAssetsDirectory)))

	return rootMux
}
