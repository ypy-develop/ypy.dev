//go:build kyse

package partials

import (
	"html/template"
	"strings"
)

@go
// ReferenceTableProps is the rows of one table.
type ReferenceTableProps struct {
	Rows []ReferenceRow
}

// ReferenceRow is one name of a table of the reference, ready to draw. The
// description and the value are markup that reference.Inline escaped before it
// put a <code> tag back.
type ReferenceRow struct {
	Name, Anchor, Extension, Direction, Guide string
	Text, ValueText                           string
	Syntax, ValueType, Type, Default          string
	Detail, Target, Value, Deprecated, Attrs  string
	Cancelable, Inherited                     bool
	Values, Modifiers                         []ReferenceValue
}

// ReferenceValue is an accepted value or a modifier, with its description drawn.
type ReferenceValue struct {
	Name string
	Text string
}

// ReferenceTable draws one table of the reference: the markup a page of
// content/ gets where it asks for <!-- reference:attributes --> and the like.
func ReferenceTable(p ReferenceTableProps) template.HTML {
	var out strings.Builder
	_ = renderPartialsReferenceTable(&out, p)
	return template.HTML(out.String())
}

@endgo

<div class="table-scroll">
	<table class="ref-table">
		<thead><tr><th scope="col">Name</th><th scope="col">What it does</th></tr></thead>
		<tbody>
			@foreach(.Rows as r)
				<tr id="{{ r.Anchor }}">
					<td class="ref-name">
						<a class="ref-anchor" href="#{{ r.Anchor }}" aria-label="Link to {{ r.Name }}"><code>{{ r.Name }}</code></a>
						@if(r.Extension != "")
							<span class="ref-tag">{{ r.Extension }}</span>
						@endif
						@if(r.Deprecated != "")
							<span class="ref-tag ref-tag-warn">deprecated</span>
						@endif
						@if(r.Cancelable)
							<span class="ref-tag">cancelable</span>
						@endif
						@if(r.Direction != "")
							<span class="ref-tag">{{ r.Direction }}</span>
						@endif
					</td>
					<td>
						<p>{!! Prose(r.Text) !!}</p>
						@if(r.Deprecated != "")
							<p class="ref-note">Deprecated. {{ r.Deprecated }}</p>
						@endif
						@if(r.Syntax != "")
							<p class="ref-meta"><span class="lbl site-muted">Syntax</span> <code>{{ r.Syntax }}</code></p>
						@endif
						@if(r.ValueType != "")
							<p class="ref-meta"><span class="lbl site-muted">Value</span> <code>{{ r.ValueType }}</code></p>
						@endif
						@if(r.Inherited)
							<p class="ref-meta"><span class="ref-tag">inherited</span></p>
						@endif
						@if(len(r.Values) > 0)
							<ul class="ref-values">
								@foreach(r.Values as v)
									<li><code>{{ v.Name }}</code> {!! Prose(v.Text) !!}</li>
								@endforeach
							</ul>
						@endif
						@if(len(r.Modifiers) > 0)
							<ul class="ref-values">
								@foreach(r.Modifiers as v)
									<li><code>{{ v.Name }}</code> {!! Prose(v.Text) !!}</li>
								@endforeach
							</ul>
						@endif
						@if(r.Type != "")
							<p class="ref-meta"><span class="lbl site-muted">Type</span> <code>{{ r.Type }}</code></p>
						@endif
						@if(r.Default != "")
							<p class="ref-meta"><span class="lbl site-muted">Default</span> <code>{{ r.Default }}</code></p>
						@endif
						@if(r.Detail != "")
							<p class="ref-meta"><span class="lbl site-muted">detail</span> <code>{{ r.Detail }}</code></p>
						@endif
						@if(r.Target != "")
							<p class="ref-meta"><span class="lbl site-muted">Fires on</span> <code>{{ r.Target }}</code></p>
						@endif
						@if(r.Value != "")
							<p class="ref-meta"><span class="lbl site-muted">Value</span> {!! Prose(r.ValueText) !!}</p>
						@endif
						@if(r.Attrs != "")
							<p class="ref-meta"><span class="lbl site-muted">Attributes</span> {{ r.Attrs }}</p>
						@endif
						@if(r.Guide != "")
							<p class="ref-meta"><a href="{{ r.Guide }}">Read more</a></p>
						@endif
					</td>
				</tr>
			@endforeach
		</tbody>
	</table>
</div>
