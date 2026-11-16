//go:build kyse

package layouts

import (
	"github.com/arandu-io/hesape/geo"
	"github.com/arandu-io/kyse/components"
	"github.com/arandu-io/kyse/icons"

	site "github.com/ypy-develop/ypy.dev/app/Site"
	partials "github.com/ypy-develop/ypy.dev/storage/framework/views/partials"
)

@go
// SiteLayout is what the chrome of this site asks of every page: the framework's
// questions about the document, and the few things this site adds -- the header
// links, which of them is current, the version the pages document, and the head
// hesape/geo built for the page: robots line, canonical address, alternates,
// social card and structured data.
type SiteLayout interface {
	view.Layout
	Nav() []site.NavItem
	IsGroup(match string) bool
	YPYVersion() string
	Robots() string
	Head() geo.Metadata
	Language() string
	GitHub() string
	CommunityURL() string
}

var _ SiteLayout = site.Chrome{}
@endgo

<!doctype html>
{{-- The theme is not declared on <html>: theme.js applies the stored mode before
     the body is parsed, so the page is painted once, in the right colours.
     This is a kyse comment, not an HTML one: it does not reach the page. --}}
<html lang="{{ .Language() }}" class="h-full">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>{{ .PageTitle() }}</title>
	<link rel="icon" href="/favicon.ico" sizes="48x48">
	<link rel="icon" href="{{ view.Asset("ypy-icon.svg") }}" type="image/svg+xml" sizes="any">
	<link rel="icon" href="/favicon-32x32.png" type="image/png" sizes="32x32">
	<link rel="icon" href="/favicon-16x16.png" type="image/png" sizes="16x16">
	<link rel="apple-touch-icon" href="/apple-touch-icon.png" sizes="180x180">
	<link rel="manifest" href="/site.webmanifest">
	<meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)">
	<meta name="theme-color" content="#000000" media="(prefers-color-scheme: dark)">
	<meta name="application-name" content="{{ .BrandName() }}">
	@if(.PageDescription() != "")
		<meta name="description" content="{{ .PageDescription() }}">
		<meta property="og:description" content="{{ .PageDescription() }}">
		<meta name="twitter:description" content="{{ .PageDescription() }}">
	@endif
	<meta name="robots" content="{{ .Robots() }}">
	<meta property="og:title" content="{{ .PageTitle() }}">
	<meta property="og:site_name" content="{{ .BrandName() }}">
	<meta property="og:type" content="website">
	<meta property="og:locale" content="{{ .Head().Locale }}">
	<meta name="twitter:title" content="{{ .PageTitle() }}">
	{{-- The rest of the head is written only for a page that may be indexed: a
	     page that asked not to be (the 404, the search, any page of a deployment
	     that is not production) has no canonical address, no card and no graph. --}}
	@if(.Head().Canonical != "")
		<link rel="canonical" href="{{ .Head().Canonical }}">
		<meta property="og:url" content="{{ .Head().Canonical }}">
		@foreach(.Head().Alternates as alt)
			<link rel="alternate" hreflang="{{ alt.Tag }}" href="{{ alt.URL }}">
		@endforeach
		@foreach(.Head().Locales as other)
			<meta property="og:locale:alternate" content="{{ other }}">
		@endforeach
		<meta property="og:image" content="{{ .Head().Image }}">
		<meta property="og:image:width" content="{{ .Head().ImageWidth }}">
		<meta property="og:image:height" content="{{ .Head().ImageHeight }}">
		<meta property="og:image:type" content="{{ .Head().ImageType }}">
		<meta property="og:image:alt" content="{{ .Head().ImageAlt }}">
		<meta name="twitter:card" content="{{ .Head().TwitterCard }}">
		<meta name="twitter:image" content="{{ .Head().Image }}">
		<meta name="twitter:image:alt" content="{{ .Head().ImageAlt }}">
		<link rel="sitemap" type="application/xml" href="/sitemap.xml">
		<link rel="describedby" type="text/plain" href="/llms.txt">
		{!! partials.Schema(.Head()) !!}
	@endif

	{{-- YPY is configured here, once, as data: a meta tag is not a script, so it
	     costs nothing against script-src 'self'. includeIndicatorCSS is off
	     because the library would inject a <style> element for it, which
	     style-src 'self' refuses; the site's stylesheet carries that rule. The
	     transitions are the page-to-page cross-fade of the stylesheet. --}}
	<meta name="ypy-config" content='{"includeIndicatorCSS":false,"transitions":true,"extensions":"ypy-head,ypy-preload"}'>

	<link rel="preload" href="{{ view.Asset("montserrat-400-600-latin.woff2") }}" as="font" type="font/woff2" crossorigin>
	<link rel="preload" href="{{ view.Asset("montserrat-600-800-latin.woff2") }}" as="font" type="font/woff2" crossorigin>
	<script src="{{ view.Asset("theme.min.js") }}"></script>
	<link rel="stylesheet" href="{{ view.Asset("app.css") }}">
	{{-- The library is the one the framework embeds, and its two extensions are the
	     project's own copies: the policy allows scripts from 'self' only, so none
	     comes from a CDN. They run in this order. --}}
	<script src="{{ view.Asset("ypy.min.js") }}" defer></script>
	<script src="{{ view.Asset("ypy-head.min.js") }}" defer></script>
	<script src="{{ view.Asset("ypy-preload.min.js") }}" defer></script>
	<script src="{{ view.Asset("app.min.js") }}" defer></script>
</head>
<body ypy-boost="true" ypy-boost:inherited="true" class="bg-background text-foreground min-h-full antialiased">
	<a class="skip-link" href="#main-content">Skip to the content</a>
	<div class="site-shell flex min-h-full w-full flex-col">
		<header class="site-header sticky top-0 z-50 w-full">
			<div class="site-header-surface bg-card w-full">
			<div class="mx-auto max-w-[1240px] px-6 max-sm:px-5">
			<div class="site-header-strip relative grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 px-3 sm:gap-4 sm:px-6 lg:grid-cols-[auto_1fr_auto] lg:gap-6">
				{!! partials.RegistrationMarks(partials.RegistrationMarksProps{}) !!}

				<div class="col-start-1 flex min-w-0 items-center gap-1 sm:gap-4">
					{{-- Below the width the row of links fits in, the pages sit in a
					     disclosure. It is <details> and not a dialog on purpose: this
					     site is served under script-src 'self' with no inline handlers,
					     and a disclosure opens with no script at all. --}}
					<details class="shrink-0 lg:hidden" name="chrome" data-mobile-navigation>
						<summary class="site-header-control text-ink/70 hover:text-ink flex size-11 cursor-pointer list-none items-center justify-center rounded-md text-[20px]" aria-label="Pages">
							{!! icons.List(icons.Props{}) !!}
						</summary>
						<div class="border-ink/10 bg-card absolute inset-x-0 top-[72px] z-[80] border-b shadow-sm">
							<nav class="mx-auto max-w-[1240px] px-6 py-2 text-[15px] font-medium" aria-label="Pages">
								@foreach(.Nav() as item)
									<a class="navlink mobile-navlink text-ink/75 hover:text-ink block py-3" href="{{ item.Href }}"
										@if(.IsGroup(item.Match))
											aria-current="page"
										@endif
										@if(item.External)
											ypy-boost="false" rel="noopener"
										@endif
									>
										<span>{{ item.Label }}</span>
									</a>
								@endforeach
							</nav>
						</div>
					</details>

					<a class="site-header-control flex min-w-0 items-center gap-2.5" href="/" aria-label="YPY, home">
						<img class="size-8 shrink-0" src="{{ view.Asset("ypy-icon.svg") }}" width="32" height="32" alt="">
						<span class="font-display whitespace-nowrap text-[17px] leading-none font-bold tracking-tight">{{ .BrandName() }}</span>
						<span class="site-version lbl hidden sm:inline">v{{ .YPYVersion() }}</span>
					</a>
				</div>

				<nav class="hidden items-center justify-center gap-0.5 text-[13.5px] leading-none font-medium lg:col-start-2 lg:flex" aria-label="Main">
					@foreach(.Nav() as item)
						<a class="navlink site-header-control text-ink/70 hover:text-ink inline-flex items-center rounded px-2.5" href="{{ item.Href }}"
							@if(.IsGroup(item.Match))
								aria-current="page"
							@endif
							@if(item.External)
								ypy-boost="false" rel="noopener"
							@endif
						>
							{!! partials.RegistrationMarks(partials.RegistrationMarksProps{Bottom: true}) !!}
							<span>{{ item.Label }}</span>
						</a>
					@endforeach
				</nav>

				<div class="col-start-2 flex items-center justify-end gap-1 sm:gap-2 lg:col-start-3">
					{{-- The search is a form that works without a script: it asks
					     /search. The page script, when it loads, answers on the spot
					     from the same index. --}}
					<form class="site-search" role="search" action="/search" method="get" data-site-search>
						<label class="sr-only" for="site-search-input">Search the documentation</label>
						<span class="site-search-icon" aria-hidden="true">{!! icons.MagnifyingGlass(icons.Props{}) !!}</span>
						<input id="site-search-input" class="site-search-input" type="search" name="q" placeholder="Search" autocomplete="off" spellcheck="false"
							aria-controls="site-search-results" aria-expanded="false" data-site-search-input>
						<kbd class="site-search-key" aria-hidden="true">/</kbd>
						<div id="site-search-results" class="site-search-results" role="listbox" aria-label="Search results" hidden data-site-search-results></div>
					</form>
					{!! components.ThemeToggle(components.ThemeToggleProps{}) !!}
					<a class="site-header-control text-ink/60 hover:text-ink hidden size-11 items-center justify-center rounded-md text-[20px] sm:inline-flex" href="{{ .GitHub() }}" ypy-boost="false" rel="noopener" aria-label="YPY on GitHub">
						{!! icons.GithubLogo(icons.Props{}) !!}
					</a>
				</div>
			</div>
			</div>
			</div>
		</header>

		<main id="main-content" tabindex="-1" class="site-main mx-auto w-full max-w-[1240px] flex-1 px-6">
			@yield('content')
		</main>

		<footer class="site-footer bg-card border-ink/10 border-t">
			<div class="mx-auto max-w-[1240px] px-6">
				<div class="col relative grid gap-10 px-8 py-14 sm:grid-cols-2 md:grid-cols-[1.4fr_repeat(3,.8fr)] max-sm:px-5">
					{!! partials.RegistrationMarks(partials.RegistrationMarksProps{Bottom: true}) !!}
					<div>
						<a class="inline-flex" href="/" aria-label="YPY, home">
							<img src="{{ view.Asset("ypy.svg") }}" width="180" height="73" loading="lazy" alt="YPY" class="footer-logo">
						</a>
						<p class="reading-copy text-ink/60 mt-4 max-w-[34ch] text-[14px] leading-relaxed">A hypermedia framework: HTML attributes ask the server for HTML and say where the answer goes. One script, no build step.</p>
					</div>
					<div>
						<p class="mb-3 text-[13px] font-semibold">Documentation</p>
						<ul class="reading-copy space-y-2 text-[13.5px]">
							<li><a class="text-ink/60 hover:text-ink" href="/docs/guides/getting-started">Getting started</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/docs">Docs</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/reference">Reference</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/docs/examples">Examples</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/docs/extensions">Extensions</a></li>
						</ul>
					</div>
					<div>
						<p class="mb-3 text-[13px] font-semibold">Server side</p>
						<ul class="reading-copy space-y-2 text-[13.5px]">
							<li><a class="text-ink/60 hover:text-ink" href="/go">Go SDK</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/docs/server">Server guides</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/docs/server/recipes/arandu">Arandu</a></li>
						</ul>
					</div>
					<div>
						<p class="mb-3 text-[13px] font-semibold">Community</p>
						<ul class="reading-copy space-y-2 text-[13.5px]">
							<li><a class="text-ink/60 hover:text-ink" href="{{ .CommunityURL() }}" ypy-boost="false" rel="noopener">Discussions</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="{{ .GitHub() }}" ypy-boost="false" rel="noopener">GitHub</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/sitemap.xml" ypy-boost="false">Sitemap</a></li>
							<li><a class="text-ink/60 hover:text-ink" href="/llms.txt" ypy-boost="false">llms.txt</a></li>
						</ul>
					</div>
				</div>
				<div class="site-footer-signature text-ink/45 border-ink/10 flex items-center justify-center border-t py-6 text-[12.5px]">
					<p class="reading-copy flex flex-wrap items-center justify-center gap-3">
						<span class="text-ink/45">Developed by</span>
						<a href="https://hyz.is" ypy-boost="false" class="text-ink/75 hover:text-ink inline-flex items-center" aria-label="HYZIS" rel="noopener">
							<span class="brandmark brandmark-hyzis" aria-hidden="true"></span>
						</a>
						<span class="text-ink/45">in</span>
						<a href="https://arandu.io" ypy-boost="false" class="text-ink/75 hover:text-ink inline-flex items-center" aria-label="Arandu" rel="noopener">
							<span class="brandmark brandmark-arandu" aria-hidden="true"></span>
						</a>
						<span class="text-ink/45">MIT licence</span>
					</p>
				</div>
			</div>
		</footer>
	</div>
</body>
</html>
