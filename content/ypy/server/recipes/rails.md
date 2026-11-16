---
order: 40
id: "RECIPE-rails"
type: "recipe"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/recipe", "lang/en", "stack/rails"]
title: "Recipe: Rails"
description: "Serve YPY from a Rails application."
---

# Recipe: Rails

A page, a form that adds an item, a fragment for YPY, a `303` for a browser without the script, a `422` for an
invalid form, `Vary: YPY-Request`, Rails' CSRF protection and Tailwind classes. There is no gem: YPY is HTTP headers
and HTML. Written for Rails 7.1 or later.

Other stacks: [Go](go.md) · [Arandu](arandu.md) · [Django](django.md) · [Laravel](laravel.md).

> A new Rails application ships with Turbo. Two libraries that take over links and forms fight each other: remove
> `turbo-rails` and its import, or mark the regions YPY owns with `data-turbo="false"`.

## What the server has to do

| The server | In Rails |
|---|---|
| tells a YPY request from a page load | `request.headers["YPY-Request"] == "true"` |
| answers a fragment to YPY and a page to everything else | `layout false` for YPY, the full layout otherwise |
| adds `Vary: YPY-Request` | a `before_action` that appends it |
| answers an invalid form with `422` | `render partial: "form", status: 422` |
| answers a valid form with `303` | `redirect_to items_path, status: :see_other` |
| passes CSRF | `form_with` adds the token; the `X-CSRF-Token` header for the rest |

## The controller

```ruby
# app/controllers/concerns/ypy_request.rb
module YpyRequest
  extend ActiveSupport::Concern

  included do
    before_action :vary_on_ypy
    layout -> { ypy_request? ? false : "application" }
    helper_method :ypy_request?
  end

  private

  def ypy_request?
    request.headers["YPY-Request"] == "true"
  end

  # A cache must never hand the fragment to a page load, or the page to a fragment.
  def vary_on_ypy
    response.headers["Vary"] = [response.headers["Vary"], "YPY-Request"].compact.join(", ")
  end
end
```

```ruby
# app/controllers/items_controller.rb
class ItemsController < ApplicationController
  include YpyRequest

  def index
    @items = Item.order(:id)
    @item = Item.new
  end

  def create
    @item = Item.new(item_params)

    if @item.save
      return redirect_to items_path, status: :see_other unless ypy_request? # no script: post, redirect, get

      @count = Item.count
      render :create # the new row for #list, and a partial that updates the counter
    else
      # 422 and the form with its errors; YPY swaps it into the form
      render partial: "form", locals: { item: @item }, status: 422
    end
  end

  private

  def item_params
    params.expect(item: [:title])
  end
end
```

(`params.expect` is Rails 8; on 7.1 and 7.2 write `params.require(:item).permit(:title)`.) The routes are
`resources :items, only: %i[index create]`.

## The views

`app/views/layouts/application.html.erb`: the layout loads the script and puts the CSRF token in a header once, for
the requests that no form backs.

```erb
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Items</title>
  <%= csrf_meta_tags %>
  <%= stylesheet_link_tag "application" %>
  <%= javascript_include_tag "ypy.min" %>
</head>
<body ypy-headers:inherited="<%= { "X-CSRF-Token" => form_authenticity_token }.to_json %>" class="bg-white text-slate-900">
  <main class="mx-auto max-w-lg p-6"><%= yield %></main>
</body>
</html>
```

`app/views/items/index.html.erb`:

```erb
<h1 class="text-2xl font-semibold">Items <span id="count" class="text-sm text-slate-500"><%= @items.size %> items</span></h1>
<%= render "form", item: @item %>
<ul id="list" class="mt-4 space-y-1">
  <%= render partial: "item", collection: @items %>
</ul>
```

`app/views/items/_form.html.erb`: used by the page and by the `422`. `form_with` adds `authenticity_token` and the
method, so the form works with no script; `ypy-error-swap="outerMorph"` says a `422` answer is the whole `<form>`.

```erb
<%= form_with model: item, id: "new", data: { turbo: false },
      html: { "ypy-post": items_path, "ypy-target": "#list", "ypy-swap": "beforeend", "ypy-error-swap": "outerMorph",
              class: "flex flex-wrap gap-2 ypy-error:ring-1 ypy-error:ring-red-500" } do |f| %>
  <%= f.text_field :title, required: true, "aria-label": "New item",
        "aria-invalid": item.errors[:title].any? ? "true" : nil,
        "aria-describedby": item.errors[:title].any? ? "title-error" : nil,
        class: "flex-1 rounded border p-2 aria-invalid:border-red-500" %>
  <%= f.button "Add", class: "rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50" %>
  <% item.errors[:title].each do |message| %>
    <p id="title-error" role="alert" class="w-full text-sm text-red-700"><%= message %></p>
  <% end %>
<% end %>
```

`app/views/items/_item.html.erb` and `create.html.erb`:

```erb
<li class="rounded border p-2"><%= item.title %></li>
```

```erb
<%= render partial: "item", locals: { item: @item } %>
<ypy-partial ypy-target="#count" ypy-swap="innerHTML"><%= @count %> items</ypy-partial>
```

ERB escapes `<%= item.title %>`; never call `raw` or `html_safe` on user text.

## CSRF

`protect_from_forgery` is on by default in `ApplicationController` and reads the token from the `authenticity_token`
field or from the `X-CSRF-Token` header.

- `form_with` writes the hidden field, so `<form ypy-post>` is covered with or without the script.
- The `ypy-headers:inherited` line in the layout covers a `ypy-delete` on a button, a `ypy-patch` on a toggle and any
  request with no form. An attribute reaches descendants only with `:inherited`.
- A failure raises `ActionController::InvalidAuthenticityToken` (`422` in production by default); the invalid form
  and a bad token must not share a status your layout treats the same way: give the token failure its own
  `rescue_from` with a `403` and a message.
- Cookies: `config.session_store :cookie_store, same_site: :lax, secure: true` in production.

## Tailwind

With `tailwindcss-rails` (v4), tell it to scan the views and add the state variants once, in
`app/assets/tailwind/application.css`:

```css
@import "tailwindcss";
@source "../../views";

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

See [Styling the states](../../guides/core-concepts.md#styling-the-states).

## Test

```ruby
# test/integration/items_test.rb
require "test_helper"

class ItemsTest < ActionDispatch::IntegrationTest
  YPY = { "YPY-Request" => "true" }.freeze

  test "ypy gets a fragment and vary" do
    post items_path, params: { item: { title: "Buy milk" } }, headers: YPY
    assert_response :success
    assert_includes response.body, "Buy milk"
    assert_includes response.body, '<ypy-partial ypy-target="#count"'
    assert_includes response.headers["Vary"], "YPY-Request"
  end

  test "an invalid form is 422 with the error" do
    post items_path, params: { item: { title: "" } }, headers: YPY
    assert_response 422
    assert_select "p[role=alert]"
  end

  test "a browser without the script gets 303" do
    post items_path, params: { item: { title: "Buy milk" } }
    assert_response :see_other
    assert_redirected_to items_path
  end

  test "user text is escaped" do
    post items_path, params: { item: { title: "<b ypy-post='/x'>" } }, headers: YPY
    assert_not_includes response.body, "<b ypy-post"
  end
end
```

Rails turns CSRF off in the test environment (`allow_forgery_protection = false`); add one test with it on and the
header present.

## Checklist

- [ ] `Vary: YPY-Request` on every action that answers a fragment and a page
- [ ] `422` with the form for invalid input, `303` for a browser without the script
- [ ] `form_with` in forms; the `X-CSRF-Token` header on the layout for the rest
- [ ] Turbo removed or kept away from YPY's regions
- [ ] no `raw` or `html_safe` on user text; no user text in an attribute name
