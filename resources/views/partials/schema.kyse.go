//go:build kyse

package partials

import (
	"html/template"

	"github.com/arandu-io/hesape/geo"
)

@go
// Schema is the page's JSON-LD block: the one graph hesape/geo and the site
// built for a page that may be indexed, in a data block the browser never runs.
// geo wrote it with json.Marshal, which escapes < > and & so it cannot close the
// block it sits in. A page that is not indexed has no graph and gets nothing.
func Schema(meta geo.Metadata) template.HTML { return meta.SchemaMarkup() }
@endgo

{{-- This view draws nothing: its body is the Go above. --}}
