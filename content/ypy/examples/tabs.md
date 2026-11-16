---
order: 180
title: Tabs
description: Tabs as plain links: the server returns the strip and the panel, YPY morphs them in and pushes the URL, so reload, sharing and Back all work.
---

# Tabs

Each tab is a link. A click asks the server for the tab strip and the panel together, and YPY morphs the answer into
the page. The server marks the selected tab, so the selection is never kept in two places. The address changes with each
tab, and the same URL opens the same tab on a full page load.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<div id="tabs">
  <div role="tablist" ypy-target:inherited="#tabs" ypy-swap:inherited="outerMorph" ypy-push-url:inherited="true">
    <a role="tab" aria-selected="true"  href="/account?tab=profile" ypy-get="/account?tab=profile">Profile</a>
    <a role="tab" aria-selected="false" href="/account?tab=billing" ypy-get="/account?tab=billing">Billing</a>
    <a role="tab" aria-selected="false" href="/account?tab=team"    ypy-get="/account?tab=team">Team</a>
  </div>
  <div role="tabpanel">Profile settings...</div>
</div>
```

The three attributes on the tab strip use `:inherited`, so every link below them gets the target, the swap style and the
history push. Without JavaScript each link is a normal link to the full page.

## The server (Go)

```go
package account

import (
	"html"
	"net/http"

	ypy "github.com/ypy-develop/go"
)

var tabs = []struct{ Key, Label string }{
	{"profile", "Profile"}, {"billing", "Billing"}, {"team", "Team"},
}

func tabsHTML(current string, body string) string {
	s := `<div id="tabs"><div role="tablist" ypy-target:inherited="#tabs" ` +
		`ypy-swap:inherited="outerMorph" ypy-push-url:inherited="true">`
	for _, t := range tabs {
		sel := "false"
		if t.Key == current {
			sel = "true"
		}
		url := "/account?tab=" + t.Key
		s += `<a role="tab" aria-selected="` + sel + `" href="` + url + `" ypy-get="` + url + `">` +
			html.EscapeString(t.Label) + `</a>`
	}
	return s + `</div><div role="tabpanel">` + body + `</div></div>`
}

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	current := r.URL.Query().Get("tab")
	body, ok := h.panels[current] // trusted HTML, built from your templates
	if !ok {
		current, body = "profile", h.panels["profile"]
	}
	markup := tabsHTML(current, body)
	if ypy.FromRequest(r).IsYPY {
		ypy.Fragment(w, markup) // only the tab block
		return
	}
	h.renderPage(w, markup) // the whole page, same URL
}
```

## How it works

- `ypy-get` on an `<a>` makes YPY take the click. A click with Ctrl, Cmd or Shift still opens the link the browser's way.
- `outerMorph` morphs the `#tabs` element into the new one. The tab you clicked keeps its focus, and only the attributes
  and content that changed are touched.
- `ypy-push-url="true"` pushes the URL of the request, `/account?tab=billing`. That URL is also a real page, so reload
  and sharing work. For a URL that differs from the request, send `ypy.PushURL(w, url)` from the handler.
- The Back button restores the earlier tab from the history. See [History](../reference/history.md).

## Variations

- Load a panel only when its tab is first used: keep the strip static and make each link target `#panel` with `innerHTML`.
  Then the server cannot update `aria-selected`; use a [`ypy-partial`](../reference/tags/ypy-partial.md) for the strip, or the morph above.
- Keep tab state out of the address bar by leaving out `ypy-push-url`.
- Add `ypy-indicator="#tabs"` to dim the block while a panel loads.

## See also

- [Updating other content](updating-other-content.md)
- [History](../reference/history.md)
- [Attribute inheritance](../reference/inheritance.md)
- [Accessibility](../guides/accessibility.md)
