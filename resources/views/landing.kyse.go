//go:build kyse

package views

import (
	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	site "github.com/ypy-develop/ypy.dev/app/Site"
)

@go
// LandingData is a folder of the manual that has no index page: it lists what is in it.
type LandingData struct {
	site.Chrome

	Title        string
	SectionTitle string
	SectionHref  string
	Pages        []*docs.Page
	Subs         []docs.Sub
	Tree         []docs.Group
}
@endgo

@extends('layouts.app')

@section('content')
	<div class="docs-layout col docs-layout-wide">
		<article class="docs-article">
			<nav class="docs-crumbs" aria-label="Breadcrumb">
				<a href="{{ .SectionHref }}">{{ .SectionTitle }}</a>
			</nav>
			<h1 class="docs-title font-display">{{ .Title }}</h1>
			<ul class="landing-list">
				@foreach(.Pages as p)
					<li class="landing-item">
						<a class="landing-title" href="{{ p.Route }}">{{ p.Title }}</a>
						<p class="reading-copy site-muted">{{ p.Description }}</p>
					</li>
				@endforeach
			</ul>
			@foreach(.Subs as s)
				<h2 class="landing-sub font-display">{{ s.Title }}</h2>
				<ul class="landing-list">
					@foreach(s.Pages as p)
						<li class="landing-item">
							<a class="landing-title" href="{{ p.Route }}">{{ p.Title }}</a>
							<p class="reading-copy site-muted">{{ p.Description }}</p>
						</li>
					@endforeach
				</ul>
			@endforeach
		</article>
	</div>
@endsection
