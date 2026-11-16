//go:build kyse

package views

import (
	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	site "github.com/ypy-develop/ypy.dev/app/Site"
)

@go
// SearchData is the search page, which answers without a script.
type SearchData struct {
	site.Chrome

	Query   string
	Results []docs.Result
}
@endgo

@extends('layouts.app')

@section('content')
	<div class="docs-layout col docs-layout-wide">
		<article class="docs-article">
			<h1 class="docs-title font-display">Search</h1>
			<form class="search-page-form" action="/search" method="get" role="search">
				<label class="sr-only" for="search-page-input">Search the documentation</label>
				<input id="search-page-input" class="search-page-input" type="search" name="q" value="{{ .Query }}" placeholder="An attribute, an event, a word" autocomplete="off" spellcheck="false">
				<button class="site-button site-button-primary" type="submit">Search</button>
			</form>
			@if(.Query != "")
				<p class="docs-meta lbl site-muted"><span>{{ len(.Results) }} results for &ldquo;{{ .Query }}&rdquo;</span></p>
				<ul class="landing-list">
					@foreach(.Results as r)
						<li class="landing-item">
							<a class="landing-title" href="{{ r.URL }}">{{ r.Title }}</a>
							<p class="lbl site-muted">{{ r.Where }}</p>
							@if(r.Description != "")
								<p class="reading-copy site-muted">{{ r.Description }}</p>
							@endif
						</li>
					@endforeach
				</ul>
			@endif
		</article>
	</div>
@endsection
