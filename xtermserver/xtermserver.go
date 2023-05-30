package xtermserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docktermj/cloudshell/xtermservice"
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
	HtmlTitle            string
	KeepalivePingTimeout int
	MaxBufferSizeBytes   int
	PathLiveness         string
	PathMetrics          string
	PathReadiness        string
	PathXtermjs          string
	ServerAddress        string
	ServerPort           int
	UrlRoutePrefix       string
}

// ----------------------------------------------------------------------------
// Interface methods
// ----------------------------------------------------------------------------

/*
The Serve method serves the XtermService over HTTP.

Input
  - ctx: A context to control lifecycle.
*/

func (xtermServer *XtermServerImpl) Serve(ctx context.Context) error {
	rootMux := http.NewServeMux()

	// Add XtermService.

	xtermService := &xtermservice.XtermServiceImpl{
		AllowedHostnames:     xtermServer.AllowedHostnames,
		Arguments:            xtermServer.Arguments,
		Command:              xtermServer.Command,
		ConnectionErrorLimit: xtermServer.ConnectionErrorLimit,
		HtmlTitle:            xtermServer.HtmlTitle,
		KeepalivePingTimeout: xtermServer.KeepalivePingTimeout,
		MaxBufferSizeBytes:   xtermServer.MaxBufferSizeBytes,
		UrlRoutePrefix:       xtermServer.UrlRoutePrefix,
	}
	xtermMux := xtermService.Handler(ctx)
	rootMux.Handle("/", xtermMux)

	// Start service.

	listenOnAddress := fmt.Sprintf("%s:%v", xtermServer.ServerAddress, xtermServer.ServerPort)
	server := http.Server{
		Addr:    listenOnAddress,
		Handler: addIncomingRequestLogging(rootMux),
	}
	fmt.Printf("starting server on interface:port '%s'...", listenOnAddress)
	return server.ListenAndServe()
}
