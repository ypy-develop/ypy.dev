//go:build kyse

package partials

import (
	"html/template"
	"strings"

	"github.com/arandu-io/kyse/icons"
)

@go
// RegistrationMarksProps selects whether a framed surface also closes with
// marks at its bottom corners. Section dividers need only the top pair; surfaces
// that explicitly close their own frame, such as the footer, need both pairs.
type RegistrationMarksProps struct {
	Bottom bool
}

// RegistrationMarks draws every registration cross through Kyse's Phosphor
// Duotone catalogue. The icons are decorative because the border intersection
// already communicates the structure visually.
func RegistrationMarks(p RegistrationMarksProps) template.HTML {
	var out strings.Builder
	_ = renderPartialsRegistrationMarks(&out, p)
	return template.HTML(out.String())
}
@endgo

<span class="mk tl">{!! icons.Plus(icons.Props{}) !!}</span>
<span class="mk tr">{!! icons.Plus(icons.Props{}) !!}</span>
@if(.Bottom)
	<span class="mk bl">{!! icons.Plus(icons.Props{}) !!}</span>
	<span class="mk br">{!! icons.Plus(icons.Props{}) !!}</span>
@endif
