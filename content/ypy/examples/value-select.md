---
order: 110
title: Dependent selects
description: Fill a model select from the make chosen in the first one: the server returns the options and owns the rule that links the two.
---

# Dependent selects

Choose a make and the list of models fills in. The page asks the server for the options, and the server owns the
rule that says which model belongs to which make.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<form ypy-post="/cars" ypy-target="#result">
  <label>Make
    <select name="make"
            ypy-get="/models"
            ypy-target="#model"
            ypy-swap="innerHTML">
      <option value="">Choose a make</option>
      <option value="audi">Audi</option>
      <option value="toyota">Toyota</option>
    </select>
  </label>

  <label>Model
    <select id="model" name="model" required>
      <option value="">Choose a make first</option>
    </select>
  </label>

  <button>Save</button>
</form>
<div id="result"></div>
```

## The server (Go)

```go
package cars

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

var models = map[string][]string{
	"audi":   {"A3", "A4", "Q5"},
	"toyota": {"Corolla", "Yaris", "Hilux"},
}

func Models(w http.ResponseWriter, r *http.Request) {
	list, ok := models[r.URL.Query().Get("make")]
	if !ok {
		ypy.Fragment(w, `<option value="">Choose a make first</option>`)
		return
	}
	var b strings.Builder
	b.WriteString(`<option value="">Choose a model</option>`)
	for _, m := range list {
		b.WriteString(`<option>` + html.EscapeString(m) + `</option>`)
	}
	ypy.Fragment(w, b.String())
}

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /models", Models)
	return ypy.Middleware(mux)
}
```

The server also checks the pair when the form is saved: a client can send any `model` it likes.

## How it works

- A `<select>` triggers on `change` by default, so `ypy-get` needs no `ypy-trigger`. It sends its own value as
  `make`.
- `ypy-target="#model"` with `ypy-swap="innerHTML"` replaces the options of the second select. The answer is only
  `<option>` elements.
- The select is inside a form that has its own `ypy-post`. A change of the select makes the `ypy-get` request of the
  select and does not submit the form. The form posts when the user presses Save.
- A change of make clears the model, because the answer brings a new list with the empty first option selected.

## Variations

- Send both values. Add `ypy-include="closest form"` to the first select if the server needs other fields of
  the form to build the list.
- Update a third place. Answer with a `<ypy-partial id="price">` next to the options (see
  [updating other content](updating-other-content.md)).
- First list from the server: load the first select with `ypy-trigger="load"` on a `<select>` of its own, or
  render it in the page.

## See also

- [`ypy-include`](../reference.md#core-attributes) and [`ypy-vals`](../reference.md#core-attributes)
- [Inline validation](inline-validation.md)
