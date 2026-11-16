---
order: 30
id: "RECIPE-django"
type: "recipe"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/recipe", "lang/en", "stack/django"]
title: "Recipe: Django"
description: "Serve YPY from a Django application."
---

# Recipe: Django

A page, a form that adds an item, a fragment for YPY, a `303` for a browser without the script, a `422` for an
invalid form, `Vary: YPY-Request`, Django's CSRF protection and Tailwind classes. There is no Django package: YPY is
HTTP headers and HTML, and Django already has what it needs. Written for Django 4.2 or later.

Other stacks: [Go](go.md) · [Arandu](arandu.md) · [Rails](rails.md) · [Laravel](laravel.md).

## What the server has to do

| The server | In Django |
|---|---|
| tells a YPY request from a page load | `request.headers.get("YPY-Request") == "true"` |
| answers a fragment to YPY and a page to everything else | one view, a template for each |
| adds `Vary: YPY-Request` | `@vary_on_headers("YPY-Request")` |
| answers an invalid form with `422` | `render(..., status=422)` with the form template |
| answers a valid form with `303` | a response with status 303 (`redirect()` is a `302`) |
| passes CSRF | `{% csrf_token %}` in forms, the `X-CSRFToken` header for the rest |

## The view

```python
# items/views.py
from django.http import HttpResponse
from django.shortcuts import render
from django.template.loader import render_to_string
from django.urls import reverse
from django.views.decorators.http import require_GET, require_POST
from django.views.decorators.vary import vary_on_headers

from .forms import ItemForm
from .models import Item


def is_ypy(request):
    return request.headers.get("YPY-Request") == "true"


def see_other(url):
    """303: the browser turns the POST into a GET, and a reload never posts again."""
    response = HttpResponse(status=303)
    response["Location"] = url
    return response


@require_GET
@vary_on_headers("YPY-Request")
def index(request):
    return render(request, "items/index.html", {"items": Item.objects.all(), "form": ItemForm()})


@require_POST
@vary_on_headers("YPY-Request")
def create(request):
    form = ItemForm(request.POST)
    if not form.is_valid():
        # 422 and the form with its errors; YPY swaps it into the form
        return render(request, "items/_form.html", {"form": form}, status=422)

    item = form.save()
    if not is_ypy(request):
        return see_other(reverse("items:index"))  # no script: post, redirect, get

    # the new row for #list, and a partial that updates the counter
    html = render_to_string("items/_created.html", {"item": item, "count": Item.objects.count()}, request=request)
    return HttpResponse(html)
```

```python
# items/forms.py
from django import forms
from .models import Item


class ItemForm(forms.ModelForm):
    class Meta:
        model = Item
        fields = ["title"]
```

## The templates

`templates/base.html`: the layout loads the script and, for requests that no form backs, puts the CSRF token in a
header once.

```html
{% load static %}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{% block title %}Items{% endblock %}</title>
  <link rel="stylesheet" href="{% static 'app.css' %}">
  <script src="{% static 'ypy.min.js' %}"></script>
</head>
<body ypy-headers:inherited='{"X-CSRFToken": "{{ csrf_token }}"}' class="bg-white text-slate-900">
  <main class="mx-auto max-w-lg p-6">{% block content %}{% endblock %}</main>
</body>
</html>
```

`templates/items/index.html`:

```html
{% extends "base.html" %}
{% block content %}
  <h1 class="text-2xl font-semibold">Items <span id="count" class="text-sm text-slate-500">{{ items|length }} items</span></h1>
  {% include "items/_form.html" %}
  <ul id="list" class="mt-4 space-y-1">
    {% for item in items %}{% include "items/_row.html" %}{% endfor %}
  </ul>
{% endblock %}
```

`templates/items/_form.html`: used by the page and by the `422`. `method` and `action` make it work with no script;
`ypy-error-swap="outerMorph"` says that a `422` answer is the whole `<form>`.

```html
<form id="new" method="post" action="{% url 'items:create' %}"
      ypy-post="{% url 'items:create' %}" ypy-target="#list" ypy-swap="beforeend" ypy-error-swap="outerMorph"
      class="flex flex-wrap gap-2 ypy-error:ring-1 ypy-error:ring-red-500">
  {% csrf_token %}
  <input id="id_title" name="title" value="{{ form.title.value|default_if_none:'' }}" required
         {% if form.title.errors %}aria-invalid="true" aria-describedby="title-error"{% endif %}
         class="flex-1 rounded border p-2 aria-invalid:border-red-500" aria-label="New item">
  <button class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50">Add</button>
  {% for error in form.title.errors %}
    <p id="title-error" role="alert" class="w-full text-sm text-red-700">{{ error }}</p>
  {% endfor %}
</form>
```

`templates/items/_row.html` and `_created.html`:

```html
<li class="rounded border p-2">{{ item.title }}</li>
```

```html
{% include "items/_row.html" %}
<ypy-partial ypy-target="#count" ypy-swap="innerHTML">{{ count }} items</ypy-partial>
```

Django escapes `{{ item.title }}`; never use `|safe` on user text.

## CSRF

Django's `CsrfViewMiddleware` is on by default and reads the token from the `csrfmiddlewaretoken` field or from the
`X-CSRFToken` header (`CSRF_HEADER_NAME`).

- `{% csrf_token %}` in the form covers `<form ypy-post>`, with or without the script.
- The `ypy-headers:inherited` line in `base.html` covers `ypy-delete` on a button, `ypy-patch` on a toggle and any
  request that has no form. An attribute reaches descendants only with `:inherited`.
- A failure is a `403`. Set `CSRF_FAILURE_VIEW` to a view that explains it; a YPY request that gets it keeps the
  page and shows the error where `ypy-error-target` says ([`ypy-error-*`](../../reference/attributes/ypy-error.md)).
- Cookies: `CSRF_COOKIE_SAMESITE = "Lax"`, `SESSION_COOKIE_SECURE = True`, `CSRF_COOKIE_SECURE = True` in production.

## Tailwind

Tell Tailwind v4 to scan the templates, or a class that appears only in a fragment is dropped from the build, and add
the two state variants once:

```css
@import "tailwindcss";
@source "../templates";

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

See [Styling the states](../../guides/core-concepts.md#styling-the-states).

## Test

```python
# items/tests.py
from django.test import TestCase
from django.urls import reverse


class CreateTests(TestCase):
    def test_ypy_gets_a_fragment_and_vary(self):
        r = self.client.post(reverse("items:create"), {"title": "Buy milk"}, headers={"YPY-Request": "true"})
        self.assertEqual(r.status_code, 200)
        self.assertContains(r, "Buy milk")
        self.assertContains(r, '<ypy-partial ypy-target="#count"')
        self.assertIn("YPY-Request", r.headers["Vary"])

    def test_invalid_form_is_422_with_the_error(self):
        r = self.client.post(reverse("items:create"), {"title": ""}, headers={"YPY-Request": "true"})
        self.assertEqual(r.status_code, 422)
        self.assertContains(r, 'role="alert"', status_code=422)

    def test_browser_without_script_gets_303(self):
        r = self.client.post(reverse("items:create"), {"title": "Buy milk"})
        self.assertEqual(r.status_code, 303)
        self.assertEqual(r["Location"], reverse("items:index"))

    def test_user_text_is_escaped(self):
        r = self.client.post(reverse("items:create"), {"title": "<b ypy-post='/x'>"}, headers={"YPY-Request": "true"})
        self.assertNotContains(r, "<b ypy-post")
```

The test client skips the CSRF check unless you build it with `Client(enforce_csrf_checks=True)`; add one test that
does, and posts with and without the header.

## Checklist

- [ ] `@vary_on_headers("YPY-Request")` on every view that answers a fragment and a page
- [ ] `422` with the form for invalid input, `303` for a browser without the script
- [ ] `{% csrf_token %}` in forms; the `X-CSRFToken` header on the layout for the rest
- [ ] no `|safe` on user text; no user text in an attribute name
- [ ] a test per audience: YPY, plain browser, invalid form
