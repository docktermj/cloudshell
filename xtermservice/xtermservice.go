package xtermservice

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"text/template"
	"time"

	"github.com/docktermj/cloudshell/internal/log"
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
	WorkingDir           string
}

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

//go:embed static/*
var static embed.FS

// ----------------------------------------------------------------------------
// Internal functions
// ----------------------------------------------------------------------------

// createRequestLog returns a logger with relevant request fields
func createRequestLog(r *http.Request, additionalFields ...map[string]interface{}) log.Logger {
	fields := map[string]interface{}{}
	if len(additionalFields) > 0 {
		fields = additionalFields[0]
	}
	if r != nil {
		fields["host"] = r.Host
		fields["remote_addr"] = r.RemoteAddr
		fields["method"] = r.Method
		fields["protocol"] = r.Proto
		fields["path"] = r.URL.Path
		fields["request_url"] = r.URL.String()
		fields["user_agent"] = r.UserAgent()
		fields["cookies"] = r.Cookies()
	}
	return log.WithFields(fields)
}

// ----------------------------------------------------------------------------
// Specific URL routes
// ----------------------------------------------------------------------------

func (xtermService *XtermServiceImpl) populateTemplate(responseWriter http.ResponseWriter, request *http.Request, filepath string) {

	templateBytes, err := static.ReadFile(filepath)
	if err != nil {
		http.Error(responseWriter, http.StatusText(500), 500)
		return
	}

	templateParsed, err := template.New("HtmlTemplate").Parse(string(templateBytes))
	if err != nil {
		http.Error(responseWriter, http.StatusText(500), 500)
		return
	}

	err = templateParsed.Execute(responseWriter, xtermService)
	if err != nil {
		http.Error(responseWriter, http.StatusText(500), 500)
		return
	}
}

func (xtermService *XtermServiceImpl) terminalJs(responseWriter http.ResponseWriter, request *http.Request) {

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

	// Add route for specific template pages.

	rootMux.HandleFunc("/xterm.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		xtermService.populateTemplate(w, r, "static/root/xterm.html")
	})

	rootMux.HandleFunc("/terminal.js", func(w http.ResponseWriter, r *http.Request) {
		xtermService.terminalJs(w, r)
		w.Header().Set("Content-Type", "text/javascript")
		xtermService.populateTemplate(w, r, "static/root/terminal.js")
	})

	// Add route for readiness probe.

	rootMux.HandleFunc(xtermService.PathReadiness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Add route for liveness probe.

	rootMux.HandleFunc(xtermService.PathLiveness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Add route for metrics.

	rootMux.Handle(xtermService.PathMetrics, promhttp.Handler())

	// Add route of static files.

	rootDir, err := fs.Sub(static, "static/root")
	if err != nil {
		panic(err)
	}
	rootMux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(rootDir))))

	return rootMux
}
