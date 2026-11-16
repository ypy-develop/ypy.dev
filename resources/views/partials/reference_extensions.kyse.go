//go:build kyse

package partials

import (
	"html/template"
	"strings"
)

@go
// ReferenceExtensionsProps is the list of official extensions.
type ReferenceExtensionsProps struct {
	Rows []ReferenceExtension
}

// ReferenceExtension is one official extension.
type ReferenceExtension struct {
	Name, Script, Guide, Adds string
}

// ReferenceExtensions draws the table of official extensions.
func ReferenceExtensions(p ReferenceExtensionsProps) template.HTML {
	var out strings.Builder
	_ = renderPartialsReferenceExtensions(&out, p)
	return template.HTML(out.String())
}
@endgo

<div class="table-scroll">
	<table class="ref-table">
		<thead><tr><th scope="col">Extension</th><th scope="col">Script</th><th scope="col">Adds</th></tr></thead>
		<tbody>
			@foreach(.Rows as e)
				<tr id="{{ e.Name }}">
					<td>
						@if(e.Guide != "")
							<a href="{{ e.Guide }}"><code>{{ e.Name }}</code></a>
						@else
							<code>{{ e.Name }}</code>
						@endif
					</td>
					<td><code>{{ e.Script }}</code></td>
					<td class="site-muted">{{ e.Adds }}</td>
				</tr>
			@endforeach
		</tbody>
	</table>
</div>
