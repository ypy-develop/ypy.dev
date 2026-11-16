---
title: "Attribute inheritance"
description: "An attribute reaches the elements below it only when you add :inherited. The rule, the table of every attribute and how to migrate a page."
order: 80
---
# Attribute inheritance

In YPY an attribute reaches the elements below it only when you say so. A bare attribute on a container stays on the
container. This page has the rule, the table of attributes, and the steps for moving a page written before YPY, where a
bare attribute on an ancestor reached its descendants.

## The rule

An ancestor passes a value down when the attribute carries the `:inherited` modifier. `:append` on a
descendant adds its own value after the inherited ones.

```html
<!-- The buttons below inherit the target and the swap style; the section does not request anything itself. -->
<section ypy-target:inherited="#results" ypy-swap:inherited="innerHTML"
         class="grid gap-4 rounded-lg border border-slate-200 p-6">
  <button ypy-get="/people" class="rounded bg-slate-900 px-3 py-2 text-white">People</button>
  <button ypy-get="/places" class="rounded bg-slate-900 px-3 py-2 text-white">Places</button>
</section>
<div id="results"></div>
```

- The nearest ancestor that passes the attribute down wins, however many levels up it is.
- An attribute on the element itself always wins over an inherited one.
- The `ypy-` and `data-ypy-` spellings are both read, and `:inherited` works with each of them.
  If `config.metaCharacter` is changed, the modifiers change with it (`ypy-target-inherited`).

### Adding with `:append`

```html
<div ypy-include:inherited=".page-fields">
  <form ypy-include:inherited:append=".form-fields">
    <!-- this button includes .page-fields, .form-fields and .button-fields -->
    <button ypy-post="/save" ypy-include:append=".button-fields">Save</button>
  </form>
</div>
```

`:inherited:append` both passes a value down and adds it to what came from above. The pieces are joined with a
comma, so `:append` fits the attributes whose value is a list (`ypy-include`, `ypy-vals`, `ypy-headers`).

### What does not inherit

```html
<!-- Not inherited: a bare attribute on a container stays on the container. -->
<div ypy-target="#results">
  <button ypy-get="/people">People</button>   <!-- swaps itself, not #results -->
</div>
```

There is no option that turns implicit inheritance on: `ypy.config.implicitInheritance` is ignored, so a page
cannot get the old behavior by accident.

## The table

One list in code describes every attribute (`test/lib/inheritance-rules.js`). The matrix test
(`test/tests/attributes/inheritance-native.js`) runs the core against that list and renders the
table below from it: the test fails when this page and the list differ.

In the `:inherited` and `:append` column, `yes` means an ancestor can pass the attribute down. `no` means the
attribute is read from the element itself only, either because it marks the element (such as `ypy-preserve`) or because
it is an event handler.

The last column is the `allowInheritance` flag the core passes when it reads the attribute. `no` means the value is
never taken from an ancestor unless `:inherited` says so. The attributes that start a request (`ypy-action`, the verbs,
`ypy-trigger`, `ypy-validate`) are read this way, so a request source never appears by accident.

<!-- inheritance-table:start -->
| Attribute | Group | `:inherited` and `:append` |
|---|---|---|
| `ypy-target` | Request | yes |
| `ypy-swap` | Request | yes |
| `ypy-select` | Request | yes |
| `ypy-select-oob` | Request | yes |
| `ypy-push-url` | Request | yes |
| `ypy-replace-url` | Request | yes |
| `ypy-confirm` | Request | yes |
| `ypy-config` | Request | yes |
| `ypy-encoding` | Request | yes |
| `ypy-sync` | Request | yes |
| `ypy-method` | Request | yes |
| `ypy-include` | Request | yes |
| `ypy-vals` | Request | yes |
| `ypy-headers` | Request | yes |
| `ypy-indicator` | Request | yes |
| `ypy-disable` | Request | yes |
| `ypy-boost` | Request | yes |
| `ypy-status:404` | Request | yes |
| `ypy-action` | What starts a request | yes |
| `ypy-get` | What starts a request | yes |
| `ypy-post` | What starts a request | yes |
| `ypy-put` | What starts a request | yes |
| `ypy-patch` | What starts a request | yes |
| `ypy-delete` | What starts a request | yes |
| `ypy-trigger` | What starts a request | yes |
| `ypy-validate` | What starts a request | yes |
| `ypy-error-target` | Phase 1 | yes |
| `ypy-error-swap` | Phase 1 | yes |
| `ypy-error-focus` | Phase 1 | yes |
| `ypy-focus` | Phase 1 | yes |
| `ypy-scroll` | Phase 1 | yes |
| `ypy-announce` | Phase 1 | yes |
| `ypy-retry` | Phase 1 | yes |
| `ypy-busy` | Phase 1 | yes |
| `ypy-swap-oob` | Element only | no |
| `ypy-preserve` | Element only | no |
| `ypy-preserve-attrs` | Element only | no |
| `ypy-on` | Element only | no |
| `ypy-history-elt` | Element only | no |
<!-- inheritance-table:end -->

The rows of group "Phase 1" are the attributes the core added later (`ypy-error-*`, `ypy-focus`,
`ypy-scroll`, `ypy-announce`, `ypy-retry`, `ypy-busy`). They are read through the same function as every other
attribute, so the same rule applies.

To change a rule, edit the list and paste the table that the failing test prints.

## Moving a page written before YPY

In markup written before YPY, a bare `ypy-target="#out"` on a container reached every request below it. YPY does not
do that, so each of those containers needs `:inherited`. `ypy-disinherit` and `ypy-inherit` are not read.

1. Run `ypy check` over the templates. It reports every container whose attribute a descendant relies on.
2. Add `:inherited` to that attribute on the container: `ypy-target="#out"` becomes `ypy-target:inherited="#out"`.
   Do it only where a descendant needs it.
3. Where `ypy-disinherit="ypy-target"` stopped a value, the value is simply no longer inherited: remove the
   `ypy-disinherit`, and write `:inherited` only on the attributes that should go down.

The full migration path is in the [migration guide](../migration/index.md).

## For extension authors

- Read an attribute only with `ypy.__attributeValue(elt, 'ypy-<name>')`; the resolver adds the other spellings. Never
  call `getAttribute('ypy-...')` yourself.
- `__attributeValue(elt, name, defaultVal, eltCollector)` follows the same rule as the core: an ancestor's value
  reaches the element only when the ancestor wrote it with `:inherited` (or `:inherited:append`). A new attribute
  needs nothing extra to stay on its own element.
- `eltCollector(value, element)` is called for each element of the chain, nearest first, instead of returning the
  value. `__findThisElements` uses it to find the elements whose value contains `this`.
