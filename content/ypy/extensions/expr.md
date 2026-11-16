---
title: "Author code without eval: the ypy-expr extension"
description: "Parses trigger filters, ypy-on handlers and js: values itself, so author code runs under script-src 'self' without eval or new Function."
order: 130
---
# Author code without `eval`: the `ypy-expr` extension

A page can write small pieces of code in its attributes: a trigger filter (`click[ctrlKey && key == 'a']`), an
`ypy-on` handler, a `js:` value in `ypy-vals`, `ypy-headers` or `ypy-confirm`. By default YPY compiles that text with
`new Function`, and a strict Content Security Policy (one without `'unsafe-eval'`) forbids it.

`ypy-expr` replaces the compiler. It parses the text itself and runs it as a set of closures, so the code works
under `script-src 'self'` and never calls `eval` or `new Function`. It understands a small, safe subset of
JavaScript, described below.

```html
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self'">
<meta name="ypy-config" content='callbacks:"all"'>
<script src="ypy.min.js"></script>
<script src="ext/ypy-expr.min.js"></script>

<input ypy-get="/search" ypy-trigger="keyup[key == 'Enter' || key == '[']" name="q"
       class="rounded border px-2 py-1">
<button ypy-on:click="this.classList.toggle('opacity-50')" class="rounded bg-slate-800 px-3 py-1 text-white">
  Dim
</button>
```

The extension is not part of `ypy.max`: it changes how every piece of author code runs, so you add it on purpose.
Load it after `ypy.js`. It
registers `expr@1`; `ypy.supports('expr@1')` answers it when [`ypy-boot`](boot.md) is loaded before it.

## What it does not change

It changes how the code runs, not whether it may run. `config.callbacks` still decides, and it is `"none"` until the
page sets it (see [Security](../index.md#security)). With `"none"` nothing compiles even with this
extension, and `ypy:security:blocked` fires as usual.

With the nonce gate of `ypy-csp` (`callbacks:"nonce"`), this extension stands aside: the gate compiles the code
itself and checks the nonce of the element every time it runs. Pages that use the gate do not need `ypy-expr`.

## The language

| Group | Supported |
|---|---|
| Values | numbers (`1`, `2.5`, `1e3`, `.5`), strings (`'a'`, `"a"`), `true`, `false`, `null`, `undefined`, lists `[1, 2]`, objects `{a: 1, 'b c': 2, d, [k]: 3}` (shorthand and computed keys) |
| Names | `this` (the element), the names the core hands to the code, then globals (`Math`, `JSON`, `document`, `window`...) |
| Access | `.`, `?.`, `[ ]`, calls `f(a, b)` and `f?.()`. A method keeps its object as `this`: `'abc'.toUpperCase()` |
| Objects | `new Date(0)` |
| Unary | `!`, `-`, `+`, `typeof`, `void`, prefix and postfix `++` and `--` |
| Binary | `+ - * / % **`, `< > <= >=`, `== != === !==`, `in`, `instanceof`, `&&`, `\|\|`, `??` |
| Conditional | `a ? b : c` |
| Assignment | `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `??=`, to a name or to a property |
| Statements | `;`, `{ }`, `let` / `const` / `var` (one name per statement), `if` / `else`, `return`, `with` |

The statements are for handlers. A trigger filter, a `js:` value and a `ypy-confirm` value are one expression.

Names. A name is looked up in the objects of an enclosing `with`, then among the values the core passes (in a
filter, the properties of the event, such as `key`, `target` or `ctrlKey`; in an `ypy-on` handler, the `detail` of the
event, the function names of the API), then among the globals. A name that is nowhere throws a `ReferenceError`, as
JavaScript does; `typeof nope` is `"undefined"`.

`ypy-on` handlers run inside a `with` over the `detail` of the event, so `ypy-on:greet="this.dataset.who = name"`
reads `detail.name`.

```html
<div ypy-on:go="let k = 1; k++; if (k == 2) this.dataset.k = k; else this.dataset.k = 'no'"></div>
```

`js:` values after `ypy-vals` are an object literal written without braces:

```html
<button ypy-get="/report" ypy-vals="js:from: this.dataset.from, count: [1, 'x'].length"
        class="rounded bg-blue-600 px-3 py-1 text-white" data-from="2026-01-01">Run</button>
```

This sends `GET /report?from=2026-01-01&count=2`. The server sees ordinary query parameters; nothing about the code
reaches it.

## What it refuses

An error carries the prefix `ypy: expr:`. For a handler it is reported with the `ypy:error` event, with the error in
`detail.error`.

Not supported (the error names the keyword or the token it stopped at): arrow functions, `function`, `class`, template strings,
`await`, `async`, `yield`, `import`, `delete`, loops (`for`, `while`), `switch`, `try`, regular expression literals,
spread, bitwise operators, the comma operator, destructuring. A `let` declares one name.

Refused on purpose, so the code cannot reach a compiler:

- the names `eval`, `Function`, `constructor`, `prototype`, `__proto__` and the `__define/lookupGetter/Setter__`
  family, as a variable, a property (`window.eval`, `o['con' + 'structor']`) or an object key;
- the compilers themselves (`eval`, `Function` and the async and generator function constructors) as a callee, as an
  argument or in `new`, so `Reflect.get(window, 'ev' + 'al')('1')` and `Reflect.apply(...)` fail with
  `the compilers are refused`;
- a string handed to `setTimeout` or `setInterval` (`a string cannot run as code`). A function is fine.

Code nested deeper than 40 levels is refused (`too deep`).

## Pitfalls

- `**` groups from left to right here: `2 ** 3 ** 2` is `64`, where JavaScript gives `512`. Use parentheses.
- Assigning to a bare name (`count = count + 1`) writes to a scratch scope that lives for one run. It does not create
  or change a global. Assign to a property (`state.count = ...`, `this.dataset.n = ...`) to change something.
- The extension does not turn code on. A page with `callbacks` unset still blocks every filter, handler and `js:`
  value.
- An arrow function in a handler fails to compile. Write the logic as a statement, or move it to a script and call
  it by name.
- Compiled code is cached by its text (256 entries). The same text compiles once.

## Checking it

`ypy.supports('expr@1')` tells a script whether the extension is loaded. The tests run a page under
`script-src 'self'` with `eval` and `new Function` forbidden, and check each feature of the table and each refusal
(`test/tests/ext/ypy-expr.js`).
