package xtermservice

import (
	"context"
	"net/http"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// The XtermServer interface...
type XtermService interface {
	Serve(ctx context.Context) (http.ServeMux, error)
	Handler(ctx context.Context) *http.ServeMux
}
