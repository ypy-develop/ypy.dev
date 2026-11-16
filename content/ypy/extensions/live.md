---
order: 200
title: "Live expressions: the ypy-live extension"
description: "Binds an element to values read from the page. The expression runs again whenever the DOM, an input or a swap may have changed those values."
---

# Live expressions: the `ypy-live` extension

Some page state never needs the server: a total that follows a checkbox, a button that stays disabled until a field is
filled, a label that mirrors an input. `ypy-live` runs a JavaScript expression when an element is processed and again
whenever the DOM, an input or a swap may have changed what the expression reads. There is no state to declare: the
expression reads the page itself.

For state that the server also reads and writes, use [`ypy-signals`](signals.md).

## Load it

```html
<meta name="ypy-config" content='{"callbacks": "all"}'>
<script src="ypy.min.js"></script>
<script src="ext/ypy-live.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-live` in it. It runs author code, so it
needs `config.callbacks` to allow code: `"all"`, or `"nonce"` together with [`ypy-csp`](csp.md). With the default
(`none`) nothing runs and `ypy:security:blocked` fires. It is listed with the other official extensions in
[Extensions](index.md).

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-live` | an expression | Runs when the element is processed and after each change that may affect it. Use it for effects. |
| `ypy-live:<name>` | an expression | Binds the value of the expression to a property of the element. Re-evaluated like `ypy-live`. |

The `<name>` of a binding is one of:

| Name | What it writes |
|---|---|
| an attribute or property (`value`, `disabled`, `aria-label`, `data-x`) | that property or attribute |
| `text` | `textContent` |
| `html` | `innerHTML` |
| `style` | the `style` of the element |
| `class` | classes, from an object (`{ big: n > 5 }`) or a string. Classes the binding does not manage are left alone. |
| `.name` | one class, on or off |
| `?name` | a boolean attribute, present or absent |

The short form `:name` works the same (`:text`, `:disabled`, `:?hidden`). It is on by default and controlled by
`config.live.bindPrefix`.

## The scope of an expression

Inside the expression these names are available:

| Name | Use |
|---|---|
| `q(selector)` | Query elements. The result reads and writes like the first element (`q('#size').value`) and has `count`, `arr()`, `attr`, `data`, `aria`, `class`. |
| `take(name, scope)`, `toggle(name, ...values)` | Move or flip a class or attribute among elements, for tabs and toggles. |
| `trigger(type, detail)` | Dispatch a custom event. |
| `insert(position, html)` | Insert HTML, and process what was added. |
| `debounce(ms)`, `forEvent(...)`, `nextFrame()` | Await a pause, an event or the next frame. |
| `attr`, `data`, `aria`, `local`, `closest` | Typed access to attributes, `data-*` and ARIA values, from the element or its nearest ancestor. |

`q()` also understands the directional words `next`, `previous`, `closest` and `first`/`last`. They only work inside
`ypy-live` and `ypy-on` expressions. The same helpers are available as `ypy.live` for use from your own scripts, plus
`ypy.live.refresh()` to run every expression again.

## Example

```html
<form ypy-post="/orders" ypy-target="#result">
  <label>Quantity <input id="qty" name="qty" type="number" min="1" value="1"></label>
  <p>Total: <output :text="'$' + (q('#qty').valueAsNumber * 12).toFixed(2)"></output></p>
  <button :disabled="!q('#qty').valueAsNumber || q('#qty').valueAsNumber < 1">Order</button>
</form>
<div id="result"></div>
```

The total and the button follow the field while the visitor types. The server does the real pricing in the handler:

```go
func order(w http.ResponseWriter, r *http.Request) {
	qty, err := strconv.Atoi(r.PostFormValue("qty"))
	if err != nil || qty < 1 {
		ypy.ValidationError(w, `<p class="error">Enter a quantity of 1 or more.</p>`)
		return
	}
	ypy.Fragment(w, fmt.Sprintf("<p>Ordered %d for $%.2f.</p>", qty, float64(qty)*12))
}
```

## Settings

| Key | Default | Meaning |
|---|---|---|
| `live.inputDebounce` | `100` | Wait after an `input` event before expressions run again. In ms or an interval string. |
| `live.bindPrefix` | `:` | Prefix of the short binding. `"ypy:"` makes `ypy:text` work. `""` or `false` turns the short form off. When Alpine.js is on the page the short form is off unless you set this. |
| `live.useDollar` | `false` | Adds `$()` as an alias of `q()`. |

## Notes

- Expressions run again after any change in the page, so keep them cheap and free of side effects, except in
  `ypy-live`. YPY logs a warning when one pass takes more than 16 ms.
- An expression may be async. An assignment that returns a promise is an error.
- A change the page cannot see (a plain variable, an external store) does not trigger a new run. Call
  `ypy.live.refresh()`.
- The extension never runs inside an element marked with `ypy-ignore`.
- Under a strict CSP the expressions need `ypy-csp` and `safeEval`, and each element needs the nonce.

## See also

[`ypy-signals`](signals.md), [`ypy-expr`](expr.md), [Alpine.js](alpine-compat.md), [Security](../reference/security.md).
