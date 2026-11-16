//go:build kyse

package errors

import site "github.com/ypy-develop/ypy.dev/app/Site"

@go
// ErrorData is what an error page is drawn from: the chrome of the site, so the
// header and the footer are the ones every page has, and the status as text,
// because it is typeset and never counted.
type ErrorData struct {
	site.Chrome

	Code    string
	Heading string
	Body    string
}
@endgo

@extends('layouts.app')

@section('content')
	<section class="border-ink/10 border-b">
		<div class="col relative flex min-h-[62vh] flex-col items-center justify-center px-8 py-24 text-center max-sm:px-5">
			<p class="font-display text-rust text-[92px] font-bold leading-none tracking-[-.03em] sm:text-[124px]">{{ .Code }}</p>
			<h1 class="font-display mt-5 text-[24px] font-bold tracking-[-.02em] sm:text-[30px]">{{ .Heading }}</h1>
			<p class="text-ink/60 mt-4 max-w-[46ch] leading-relaxed">{{ .Body }}</p>
			<div class="mt-8 flex flex-wrap items-center justify-center gap-3">
				<a class="site-button site-button-primary" href="/">Home</a>
				<a class="site-button site-button-secondary" href="/docs">Docs</a>
			</div>
		</div>
	</section>
@endsection
