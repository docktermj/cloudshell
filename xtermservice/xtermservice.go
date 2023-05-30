package xtermservice

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"reflect"
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

	// Add route to metrics.

	rootMux.Handle(xtermService.PathMetrics, promhttp.Handler())

	// Add route to xterm.js assets.

	// depenenciesDirectory := path.Join(xtermService.WorkingDir, "./node_modules")
	// rootMux.Handle("/assets", http.FileServer(http.Dir(depenenciesDirectory)))

	// nodeModulesDir, err := fs.Sub(static, "static/node_modules")
	// if err != nil {
	// 	panic(err)
	// }
	// rootMux.Handle("/assets", http.StripPrefix("/", http.FileServer(http.FS(nodeModulesDir))))

	// Add route to website.

	// publicAssetsDirectory := path.Join(xtermService.WorkingDir, "./public")
	// rootMux.Handle("/", http.FileServer(http.Dir(publicAssetsDirectory)))

	rootDir, err := fs.Sub(static, "static/root")
	if err != nil {
		panic(err)
	}

	// assetsDir, err := fs.Sub(static, "static/root/node_modules")
	// if err != nil {
	// 	panic(err)
	// }

	fs.WalkDir(rootDir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		fmt.Printf(">>>>>> path: %s\n", path)
		return nil
	})

	fmt.Printf(">>>>>> http.FS: %s\n", reflect.TypeOf(http.FS(rootDir)))
	fmt.Printf(">>>>>> http.FileServer: %s\n", reflect.TypeOf(http.FileServer(http.FS(rootDir))))

	// rootMux.Handle("/assets/", http.StripPrefix("/", http.FileServer(http.FS(assetsDir)))) BAD
	// rootMux.Handle("/assets/", http.FileServer(http.FS(assetsDir)))

	rootMux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(rootDir))))

	return rootMux
}
