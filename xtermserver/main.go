package xtermserver

import (
	"context"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// The XtermServer interface...
type XtermServer interface {
	Serve(ctx context.Context) error
}
