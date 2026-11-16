//go:build kyse

package partials

import "html/template"

@go
// Prose passes through markup that was made safe before it got here: the body of
// a page, which github.com/hyz-is/arandu-markdown wrote through its closed
// allowlist, and a description of the reference, which Inline escaped before it
// put a <code> tag back.
//
// Raw output is earned by a function that returns template.HTML over a value
// that was sanitised upstream, and this is that function. A string that did not
// come from one of those two does not belong here.
func Prose(html string) template.HTML {
	return template.HTML(html)
}
@endgo

{{-- This view draws nothing. The package's other files render markup; this one
     exists for the function above, and kyse compiles a file per view, so the
     body is deliberately empty. --}}
