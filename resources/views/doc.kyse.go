//go:build kyse

package views

import (
	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	partials "github.com/ypy-develop/ypy.dev/storage/framework/views/partials"
)

@go
// DocData is what the documentation controller hands a Markdown page.
type DocData struct {
	site.Chrome

	Page *docs.Page
	Body string
	// Tree is the sidebar of the section the page belongs to, in its language.
	Tree []docs.Group
	// SectionTitle names the section, and SectionHref is where it begins.
	SectionTitle string
	SectionHref  string
	Prev         *docs.Page
	Next         *docs.Page
	// Twin is the same page in the other language, nil when there is none.
	Twin *docs.Page
}

// Holds says whether a sidebar group contains the page being shown. The group
// with the current page is open; the others are closed.
func (d DocData) Holds(g docs.Group) bool {
	for _, p := range g.Pages {
		if p == d.Page {
			return true
		}
	}
	for _, s := range g.Subs {
		for _, p := range s.Pages {
			if p == d.Page {
				return true
			}
		}
	}
	return false
}

// IsHere says whether a sidebar entry is the page being shown.
func (d DocData) IsHere(p *docs.Page) bool { return p == d.Page }

// Entries are the headings the contents lists: the sections and their
// subsections.
func (d DocData) Entries() int { return len(d.Page.Headings) }
@endgo

@extends('layouts.app')

@section('content')
	<div class="docs-layout col">
		<aside class="docs-sidebar" aria-label="Documentation">
			{{-- Open in the markup, so it is a plain list with no script. The page script
			     closes it below the width where the sidebar is a column of its own. --}}
			<details class="docs-menu" open>
				<summary class="docs-menu-summary">
					<span>{{ .SectionTitle }} menu</span>
				</summary>
				<nav class="docs-nav" aria-label="{{ .SectionTitle }}">
					@foreach(.Tree as g)
						@if(g.Key == "")
							<ul class="docs-nav-list">
								@foreach(g.Pages as p)
									<li>
										<a href="{{ p.Route }}" class="docs-nav-link"
											@if(.IsHere(p))
												aria-current="page"
											@endif
										>{{ p.Title }}</a>
									</li>
								@endforeach
							</ul>
						@else
							<details class="docs-group"
								@if(.Holds(g))
									open
								@endif
							>
								<summary class="docs-group-title">{{ g.Title }}</summary>
								<ul class="docs-nav-list">
									@foreach(g.Pages as p)
										<li>
										<a href="{{ p.Route }}" class="docs-nav-link"
											@if(.IsHere(p))
												aria-current="page"
											@endif
										>{{ p.Title }}</a>
									</li>
									@endforeach
									@foreach(g.Subs as s)
										<li class="docs-sub">
											<p class="docs-sub-title">{{ s.Title }}</p>
											<ul class="docs-nav-list">
												@foreach(s.Pages as p)
													<li>
										<a href="{{ p.Route }}" class="docs-nav-link"
											@if(.IsHere(p))
												aria-current="page"
											@endif
										>{{ p.Title }}</a>
									</li>
												@endforeach
											</ul>
										</li>
									@endforeach
								</ul>
							</details>
						@endif
					@endforeach
				</nav>
			</details>
		</aside>

		<article class="docs-article" data-pagefind-body>
			<nav class="docs-crumbs" aria-label="Breadcrumb">
				@foreach(.Crumbs as crumb)
					@if(crumb.Href == .Page.Route)
						<span aria-current="page">{{ crumb.Name }}</span>
					@else
						<a href="{{ crumb.Href }}">{{ crumb.Name }}</a>
						<span aria-hidden="true">/</span>
					@endif
				@endforeach
			</nav>

			<h1 class="docs-title font-display">{{ .Page.Title }}</h1>
			<p class="docs-meta lbl site-muted">
				@if(.Page.Minutes > 0)
					<span>{{ .Page.Minutes }} min read</span>
				@endif
				@if(.Page.Updated != "")
					<span>Updated {{ .Page.Updated }}</span>
				@endif
				@if(.Twin != nil)
					<a class="text-rust" href="{{ .Twin.Route }}" ypy-boost="false" hreflang="{{ .Twin.Lang }}">{{ .Twin.Lang }}</a>
				@endif
			</p>

			<div class="prose">{!! partials.Prose(.Body) !!}</div>

			<nav class="docs-pager" aria-label="Previous and next">
				@if(.Prev != nil)
					<a class="docs-pager-link" href="{{ .Prev.Route }}" rel="prev">
						<span class="lbl site-muted">Previous</span>
						<span class="docs-pager-title">{{ .Prev.Title }}</span>
					</a>
				@else
					<span></span>
				@endif
				@if(.Next != nil)
					<a class="docs-pager-link docs-pager-next" href="{{ .Next.Route }}" rel="next">
						<span class="lbl site-muted">Next</span>
						<span class="docs-pager-title">{{ .Next.Title }}</span>
					</a>
				@endif
			</nav>
		</article>

		<aside class="docs-toc" aria-label="On this page">
			@if(len(.Page.Headings) > 0)
				<p class="docs-toc-title lbl site-muted">On this page</p>
				<ul>
					@foreach(.Page.Headings as h)
						<li class="docs-toc-item" data-level="{{ h.Level }}"><a href="#{{ h.ID }}">{{ h.Text }}</a></li>
					@endforeach
				</ul>
			@endif
		</aside>
	</div>
@endsection
