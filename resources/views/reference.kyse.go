//go:build kyse

package views

import (
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	partials "github.com/ypy-develop/ypy.dev/storage/framework/views/partials"
)

@go
// ReferenceData is the reference: an index of the tables when Kind is empty,
// and one table otherwise.
type ReferenceData struct {
	site.Chrome

	Version string
	Kind    string
	Title   string
	Intro   string
	// Tables are the pages of the reference, with the count of each.
	Tables []Table
	// Rows are the rows of the table being shown.
	Rows []Row
	// Extensions is the list of official extensions, on the extensions page.
	Extensions []ExtensionRow
}

// Table is a link to one table of the reference.
type Table struct {
	Title string
	Href  string
	Count int
	Body  string
}

// Row is one name of a table, ready to draw.
type Row struct {
	reference.Item

	// Text is the description with its `code` spans drawn.
	Text string
	// Guide is the page of the manual that explains the name, "" when none does.
	Guide string
	// ValueText is the value of a header, drawn.
	ValueText string
	// AttrList is the attributes of an element, as words.
	AttrList string
	// Values and Modifiers are the accepted values, with their descriptions drawn.
	ValueList []ValueRow
	ModList   []ValueRow
}

// ValueRow is an accepted value or a modifier.
type ValueRow struct {
	Name string
	Text string
}

// ExtensionRow is one official extension.
type ExtensionRow struct {
	reference.Extension

	Guide string
	Names int
	// Adds says in words what the extension adds: "3 attributes, 2 events".
	Adds string
}
@endgo

@extends('layouts.app')

@section('content')
	<div class="docs-layout col docs-layout-wide">
		<article class="docs-article docs-article-wide">
			<nav class="docs-crumbs" aria-label="Breadcrumb">
				<a href="/reference">Reference</a>
				@if(.Kind != "")
					<span aria-hidden="true">/</span>
					<span>{{ .Title }}</span>
				@endif
			</nav>
			<h1 class="docs-title font-display">{{ .Title }}</h1>
			<p class="docs-meta lbl site-muted"><span>YPY {{ .Version }}</span><span>generated from the library's own name tables</span></p>
			<p class="reading-copy site-muted ref-intro">{!! partials.Prose(.Intro) !!}</p>

			@if(.Kind == "")
				<ul class="repository-grid mt-8">
					@foreach(.Tables as t)
						<li class="repository-card">
							<div class="repository-card-head">
								<a href="{{ t.Href }}">{{ t.Title }} <span class="repository-arrow" aria-hidden="true">&rarr;</span></a>
								<span>{{ t.Count }}</span>
							</div>
							<p class="reading-copy">{{ t.Body }}</p>
						</li>
					@endforeach
				</ul>
			@elseif(.Kind == "extensions")
				<div class="table-scroll">
					<table class="ref-table">
						<thead><tr><th scope="col">Extension</th><th scope="col">Script</th><th scope="col">Adds</th></tr></thead>
						<tbody>
							@foreach(.Extensions as e)
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
			@else
				<div class="table-scroll">
					<table class="ref-table">
						<thead>
							<tr>
								<th scope="col">Name</th>
								<th scope="col">What it does</th>
								<th scope="col">Guide</th>
							</tr>
						</thead>
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
										<p>{!! partials.Prose(r.Text) !!}</p>
										@if(r.Deprecated != "")
											<p class="ref-note">Deprecated. {{ r.Deprecated }}</p>
										@endif
										@if(r.Syntax != "")
											<p class="ref-meta"><span class="lbl site-muted">Syntax</span> <code>{{ r.Syntax }}</code></p>
										@endif
										@if(r.ValueType != "")
											<p class="ref-meta"><span class="lbl site-muted">Value</span> <code>{{ r.ValueType }}</code></p>
										@if(r.Inherited)
											<p class="ref-meta"><span class="ref-tag">inherited</span></p>
										@endif
										@endif
										@if(len(r.ValueList) > 0)
											<ul class="ref-values">
												@foreach(r.ValueList as v)
													<li><code>{{ v.Name }}</code> {!! partials.Prose(v.Text) !!}</li>
												@endforeach
											</ul>
										@endif
										@if(len(r.ModList) > 0)
											<ul class="ref-values">
												@foreach(r.ModList as v)
													<li><code>{{ v.Name }}</code> {!! partials.Prose(v.Text) !!}</li>
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
											<p class="ref-meta"><span class="lbl site-muted">Value</span> {!! partials.Prose(r.ValueText) !!}</p>
										@endif
										@if(len(r.Attrs) > 0)
											<p class="ref-meta"><span class="lbl site-muted">Attributes</span> {{ r.AttrList }}</p>
										@endif
									</td>
									<td class="ref-guide">
										@if(r.Guide != "")
											<a href="{{ r.Guide }}">Guide</a>
										@else
											<span class="site-muted">&mdash;</span>
										@endif
									</td>
								</tr>
							@endforeach
						</tbody>
					</table>
				</div>
			@endif
		</article>
	</div>
@endsection
