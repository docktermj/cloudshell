package xtermservice

import (
	"context"
	"net/http"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// The XtermService interface...
type XtermService interface {
	Handler(ctx context.Context) *http.ServeMux
}
