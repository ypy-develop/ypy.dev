//go:build kyse

package views

import (
	site "github.com/ypy-develop/ypy.dev/app/Site"
	partials "github.com/ypy-develop/ypy.dev/storage/framework/views/partials"
)

@go
// HomeData is what HomeController.Index hands the home page.
type HomeData struct {
	site.Chrome

	// Version is YPY's version, "0.1.1", read from the file YPY ships.
	Version string
	// Scripts are the two published addresses of the script, as full <script> tags.
	JsDelivrTag string
	YpyTag      string
	// NpmInstall is the command that adds the package from npm.
	NpmInstall string
	// GoGet is the command that adds the Go SDK.
	GoGet string
	// Cards are the doors of the site.
	Cards []Card
	// Example is the minimal page, shown as code.
	Example string
	// Counts are what the reference holds, for the cards.
	Attributes int
	Events     int
	Extensions int
}

// Card is a link to a part of the site.
type Card struct {
	Title string
	Body  string
	Href  string
	// External marks a link that leaves the site.
	External bool
}
@endgo

@extends('layouts.app')

@section('content')
	<section class="site-hero border-ink/10 border-b">
		<div class="mx-auto max-w-[1240px] px-6 max-sm:px-5">
		<div class="hero-grid col">
			<div class="hero-copy">
				<p class="hero-badge lbl">
					<span class="bg-teal h-1.5 w-1.5 rounded-full" aria-hidden="true"></span>
					v{{ .Version }} &middot; MIT
				</p>

				<h1 class="hero-title font-display">HTML is the API.</h1>

				<p class="hero-body reading-copy">
					YPY is a hypermedia framework. An attribute on an element says which request to make and where
					the answer goes. The server answers with HTML. One script, no build step, no dependency.
				</p>

				<div class="mt-6 flex flex-wrap items-center gap-3">
					<a class="site-button site-button-primary" href="/docs#installing">Get started</a>
					<a class="site-button site-button-secondary" href="/reference">Reference</a>
				</div>
			</div>

			<figure class="hero-stage">
				<div class="hero-emblem hero-logo">
					<img src="{{ view.Asset("ypy.svg") }}" width="371" height="151" fetchpriority="high" alt="YPY">
				</div>
				<div class="command-bar mt-5 w-full">
					<span class="text-teal" aria-hidden="true">$</span>
					<code class="flex-1 truncate">{{ .GoGet }}</code>
					<button type="button" class="command-copy" data-copy data-copy-text="{{ .GoGet }}" aria-label="Copy the install command">
						<span data-copy-label="idle">Copy</span>
						<span data-copy-label="done">Copied</span>
					</button>
				</div>
				<figcaption class="reading-copy site-muted mt-2 w-full font-mono text-[11.5px]">
					The server side, in Go. The browser side is one script tag, below.
				</figcaption>
			</figure>
		</div>
		</div>
	</section>

	<section id="install" class="story-section border-ink/10 border-b">
		{!! partials.RegistrationMarks(partials.RegistrationMarksProps{}) !!}
		<div class="col relative grid gap-10 px-8 py-16 lg:grid-cols-[.9fr_1.1fr] lg:py-20 max-sm:px-5">
			<div class="section-copy">
				<div class="section-kicker">
					<span class="section-number" aria-hidden="true">01</span>
					<p class="lbl text-rust">Install</p>
				</div>
				<h2 class="section-title">One script tag.</h2>
				<p class="reading-copy site-muted mt-5 max-w-[47ch] leading-relaxed">
					Put it in the page, from either address, or add the package from npm. There is nothing to compile and nothing to configure
					to start; the same file works in a static page and in an application.
				</p>
				<p class="reading-copy site-muted mt-4 max-w-[47ch] text-[13.5px] leading-relaxed">
					The Go SDK answers fragments and forms, and tells a YPY request from a browser one.
					Other stacks use the same headers: see <a class="text-rust underline" href="/docs#server-side">server side</a>.
				</p>
			</div>

			<div class="grid min-w-0 gap-4">
				<figure class="code-window code-window-compact">
					<div class="code-window-bar">
						<span class="code-window-dots" aria-hidden="true"><i></i><i></i><i></i></span>
						<span class="code-window-title">jsDelivr</span>
						<span class="code-window-action">
							<button type="button" class="code-window-copy" data-copy data-copy-text="{{ .JsDelivrTag }}" aria-label="Copy the script tag">
								<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span>
							</button>
						</span>
					</div>
					<pre><code>{{ .JsDelivrTag }}</code></pre>
				</figure>
				<figure class="code-window code-window-compact">
					<div class="code-window-bar">
						<span class="code-window-dots" aria-hidden="true"><i></i><i></i><i></i></span>
						<span class="code-window-title">cdn.ypy.dev</span>
						<span class="code-window-action">
							<button type="button" class="code-window-copy" data-copy data-copy-text="{{ .YpyTag }}" aria-label="Copy the script tag">
								<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span>
							</button>
						</span>
					</div>
					<pre><code>{{ .YpyTag }}</code></pre>
				</figure>
				<figure class="code-window code-window-compact">
					<div class="code-window-bar">
						<span class="code-window-dots" aria-hidden="true"><i></i><i></i><i></i></span>
						<span class="code-window-title">npm</span>
						<span class="code-window-action">
							<button type="button" class="code-window-copy" data-copy data-copy-text="{{ .NpmInstall }}" aria-label="Copy the command">
								<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span>
							</button>
						</span>
					</div>
					<pre><code>{{ .NpmInstall }}</code></pre>
				</figure>
				<figure class="code-window code-window-compact">
					<div class="code-window-bar">
						<span class="code-window-dots" aria-hidden="true"><i></i><i></i><i></i></span>
						<span class="code-window-title">Go SDK</span>
						<span class="code-window-action">
							<button type="button" class="code-window-copy" data-copy data-copy-text="{{ .GoGet }}" aria-label="Copy the command">
								<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span>
							</button>
						</span>
					</div>
					<pre><code>{{ .GoGet }}</code></pre>
				</figure>
			</div>
		</div>
	</section>

	<section id="example" class="story-section border-ink/10 border-b">
		{!! partials.RegistrationMarks(partials.RegistrationMarksProps{}) !!}
		<div class="col relative grid gap-10 px-8 py-16 lg:grid-cols-[.9fr_1.1fr] lg:items-center lg:py-20 max-sm:px-5">
			<div class="section-copy">
				<div class="section-kicker">
					<span class="section-number" aria-hidden="true">02</span>
					<p class="lbl text-rust">A page</p>
				</div>
				<h2 class="section-title">Read it aloud.</h2>
				<p class="reading-copy site-muted mt-5 max-w-[47ch] leading-relaxed">
					On click, GET <code>/demo/hello</code>, put the answer into <code>#demo-out</code>. That is the
					whole contract, and the button on the right runs it against this site.
				</p>
				<a class="text-rust mt-6 inline-flex min-h-11 items-center font-semibold" href="/docs#introduction">Introduction <span class="ml-2" aria-hidden="true">&rarr;</span></a>
			</div>

			<div class="grid min-w-0 gap-4">
				<figure class="code-window code-window-compact">
					<div class="code-window-bar">
						<span class="code-window-dots" aria-hidden="true"><i></i><i></i><i></i></span>
						<span class="code-window-title">index.html</span>
						<span class="code-window-action">
							<button type="button" class="code-window-copy" data-copy data-copy-text="{{ .Example }}" aria-label="Copy the example">
								<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span>
							</button>
						</span>
					</div>
					<pre><code>{{ .Example }}</code></pre>
				</figure>
				<div class="demo-panel">
					<button type="button" class="site-button site-button-primary" ypy-get="/demo/hello" ypy-target="#demo-out" ypy-indicator="#demo-out">Say hello</button>
					<p id="demo-out" class="demo-out reading-copy" aria-live="polite">The answer lands here.</p>
				</div>
			</div>
		</div>
	</section>

	<section class="story-section border-ink/10 border-b">
		{!! partials.RegistrationMarks(partials.RegistrationMarksProps{}) !!}
		<div class="col relative px-8 py-16 lg:py-20 max-sm:px-5">
			<div class="section-kicker">
				<span class="section-number" aria-hidden="true">03</span>
				<p class="lbl text-rust">Where to go</p>
			</div>
			<h2 class="section-title max-w-[24ch]">{{ .Attributes }} attributes, {{ .Events }} events, {{ .Extensions }} extensions.</h2>

			<ul class="repository-grid mt-10">
				@foreach(.Cards as c)
					<li class="repository-card">
						<div class="repository-card-head">
							@if(c.External)
								<a href="{{ c.Href }}" ypy-boost="false" rel="noopener">{{ c.Title }} <span class="repository-arrow" aria-hidden="true">&nearr;</span></a>
							@else
								<a href="{{ c.Href }}">{{ c.Title }} <span class="repository-arrow" aria-hidden="true">&rarr;</span></a>
							@endif
						</div>
						<p class="reading-copy">{{ c.Body }}</p>
					</li>
				@endforeach
			</ul>
		</div>
	</section>
@endsection
