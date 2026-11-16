---
order: 50
id: "RECIPE-laravel"
type: "recipe"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/recipe", "lang/en", "stack/laravel"]
title: "Recipe: Laravel"
description: "Serve YPY from a Laravel application."
---

# Recipe: Laravel

A page, a form that adds an item, a fragment for YPY, a `303` for a browser without the script, a `422` for an
invalid form, `Vary: YPY-Request`, Laravel's CSRF protection and Tailwind classes. There is no package: YPY is HTTP
headers and HTML. Written for Laravel 11 or later, with Blade.

Other stacks: [Go](go.md) · [Arandu](arandu.md) · [Django](django.md) · [Rails](rails.md).

## What the server has to do

| The server | In Laravel |
|---|---|
| tells a YPY request from a page load | `$request->header('YPY-Request') === 'true'` |
| answers a fragment to YPY and a page to everything else | one controller action, a Blade view for each |
| adds `Vary: YPY-Request` | a small middleware on the `web` group |
| answers an invalid form with `422` | `response()->view('items._form', [...], 422)` |
| answers a valid form with `303` | `redirect()->route('items.index', [], 303)` (a plain `redirect()` is `302`) |
| passes CSRF | `@csrf` in forms; the `X-CSRF-TOKEN` header for the rest |

## Why not `$request->validate()`

For a browser request, `validate()` redirects back with the errors in the session (`302`). That keeps working for a
browser with no script, but a YPY request would follow the redirect and swap a whole page. Validate by hand with the
`Validator` and answer the form with `422` instead; the form is still the same Blade view.

## The middleware

```php
<?php
// app/Http/Middleware/VaryOnYpy.php
namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

// A cache must never hand the fragment to a page load, or the page to a fragment.
class VaryOnYpy
{
    public function handle(Request $request, Closure $next): Response
    {
        $response = $next($request);
        $response->setVary('YPY-Request', false); // false: add to the existing Vary, do not replace it

        return $response;
    }
}
```

```php
// bootstrap/app.php
->withMiddleware(function (Middleware $middleware) {
    $middleware->appendToGroup('web', \App\Http\Middleware\VaryOnYpy::class);
})
```

## The controller

```php
<?php
// app/Http/Controllers/ItemController.php
namespace App\Http\Controllers;

use App\Models\Item;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

class ItemController extends Controller
{
    private function isYpy(Request $request): bool
    {
        return $request->header('YPY-Request') === 'true';
    }

    public function index()
    {
        return view('items.index', ['items' => Item::orderBy('id')->get()]);
    }

    public function store(Request $request)
    {
        $validator = Validator::make($request->all(), ['title' => ['required', 'string', 'max:120']]);

        if ($validator->fails()) {
            // 422 and the form with its errors; YPY swaps it into the form
            return response()->view('items._form', [
                'errors' => $validator->errors(),
                'title' => (string) $request->input('title', ''),
            ], 422);
        }

        $item = Item::create($validator->validated());

        if (! $this->isYpy($request)) {
            return redirect()->route('items.index', [], 303); // no script: post, redirect, get
        }

        // the new row for #list, and a partial that updates the counter
        return response()->view('items._created', ['item' => $item, 'count' => Item::count()]);
    }
}
```

```php
// routes/web.php
Route::get('/items', [ItemController::class, 'index'])->name('items.index');
Route::post('/items', [ItemController::class, 'store'])->name('items.store');
```

## The views

`resources/views/layouts/app.blade.php`: the layout loads the script and puts the CSRF token in a header once, for
the requests that no form backs.

```blade
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="csrf-token" content="{{ csrf_token() }}">
  <title>Items</title>
  @vite('resources/css/app.css')
  <script src="{{ asset('ypy.min.js') }}"></script>
</head>
<body ypy-headers:inherited='{"X-CSRF-TOKEN": "{{ csrf_token() }}"}' class="bg-white text-slate-900">
  <main class="mx-auto max-w-lg p-6">@yield('content')</main>
</body>
</html>
```

`resources/views/items/index.blade.php`:

```blade
@extends('layouts.app')
@section('content')
  <h1 class="text-2xl font-semibold">Items <span id="count" class="text-sm text-slate-500">{{ $items->count() }} items</span></h1>
  @include('items._form', ['errors' => new \Illuminate\Support\MessageBag, 'title' => ''])
  <ul id="list" class="mt-4 space-y-1">
    @foreach ($items as $item) @include('items._row', ['item' => $item]) @endforeach
  </ul>
@endsection
```

`resources/views/items/_form.blade.php`: used by the page and by the `422`. `method` and `action` make it work with no
script; `ypy-error-swap="outerMorph"` says a `422` answer is the whole `<form>`.

```blade
<form id="new" method="post" action="{{ route('items.store') }}"
      ypy-post="{{ route('items.store') }}" ypy-target="#list" ypy-swap="beforeend" ypy-error-swap="outerMorph"
      class="flex flex-wrap gap-2 ypy-error:ring-1 ypy-error:ring-red-500">
  @csrf
  <input id="title" name="title" value="{{ $title }}" required aria-label="New item"
         @if ($errors->has('title')) aria-invalid="true" aria-describedby="title-error" @endif
         class="flex-1 rounded border p-2 aria-invalid:border-red-500">
  <button class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50">Add</button>
  @error('title')
    <p id="title-error" role="alert" class="w-full text-sm text-red-700">{{ $message }}</p>
  @enderror
</form>
```

`_row.blade.php` and `_created.blade.php`:

```blade
<li class="rounded border p-2">{{ $item->title }}</li>
```

```blade
@include('items._row', ['item' => $item])
<ypy-partial ypy-target="#count" ypy-swap="innerHTML">{{ $count }} items</ypy-partial>
```

`{{ }}` escapes. Never use `{!! !!}` on user text.

## CSRF

The `web` group verifies the token. It reads the `_token` field, or the `X-CSRF-TOKEN` header (and `X-XSRF-TOKEN`).

- `@csrf` in the form covers `<form ypy-post>`, with or without the script.
- The `ypy-headers:inherited` line in the layout covers a `ypy-delete` on a button, a `ypy-patch` on a toggle and any
  request with no form. An attribute reaches descendants only with `:inherited`.
- An expired token is a `419 Page Expired`. Customize the page in `bootstrap/app.php` (`withExceptions`) so a YPY
  request that gets it explains it, and a plain browser is sent back to the form.
- Cookies: `SESSION_SECURE_COOKIE=true` and `SESSION_SAME_SITE=lax` in production.

## Tailwind

Tailwind v4 detects the sources of a Laravel application by itself, but say so, and add the two state variants once,
in `resources/css/app.css`:

```css
@import "tailwindcss";
@source "../views";

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

See [Styling the states](../../guides/core-concepts.md#styling-the-states).

## Test

```php
<?php
// tests/Feature/ItemTest.php
namespace Tests\Feature;

use Tests\TestCase;
use Illuminate\Foundation\Testing\RefreshDatabase;

class ItemTest extends TestCase
{
    use RefreshDatabase;

    private array $ypy = ['YPY-Request' => 'true'];

    public function test_ypy_gets_a_fragment_and_vary(): void
    {
        $response = $this->withHeaders($this->ypy)->post('/items', ['title' => 'Buy milk']);

        $response->assertOk()->assertSee('Buy milk')->assertSee('<ypy-partial ypy-target="#count"', false);
        $this->assertStringContainsString('YPY-Request', $response->headers->get('Vary'));
    }

    public function test_an_invalid_form_is_422_with_the_error(): void
    {
        $this->withHeaders($this->ypy)->post('/items', ['title' => ''])
            ->assertStatus(422)->assertSee('role="alert"', false);
    }

    public function test_a_browser_without_the_script_gets_303(): void
    {
        $this->post('/items', ['title' => 'Buy milk'])->assertStatus(303)->assertRedirect('/items');
    }

    public function test_user_text_is_escaped(): void
    {
        $this->withHeaders($this->ypy)->post('/items', ['title' => "<b ypy-post='/x'>"])
            ->assertDontSee('<b ypy-post', false);
    }
}
```

The CSRF middleware lets every request through while the application runs unit tests; add one test that enables it and posts with and without the header.

## Checklist

- [ ] the `VaryOnYpy` middleware is on the `web` group
- [ ] `422` with the form for invalid input, `303` for a browser without the script
- [ ] `@csrf` in forms; the `X-CSRF-TOKEN` header on the layout for the rest
- [ ] no `{!! !!}` on user text; no user text in an attribute name
- [ ] a test per audience: YPY, plain browser, invalid form
