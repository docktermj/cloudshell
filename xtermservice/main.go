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
	Handler(ctx context.Context) *http.ServeMux
}
