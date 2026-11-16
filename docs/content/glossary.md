# Glossary of the YPY documentation

One name per thing. Use the official term; do not use the listed synonyms.

| Term | Meaning | Do not write |
|---|---|---|
| fragment | a piece of HTML the server sends for one swap | partial page, snippet, component (unless it is a Web Component) |
| target | the element a response is put into (`ypy-target`) | destination, container |
| swap | putting the response into the target (`ypy-swap`) | replace, inject, render (as a verb for this step) |
| morph | the default swap: merges new HTML into existing nodes | diff, patch (patch is reserved for streams and signals) |
| trigger | the event that starts a request (`ypy-trigger`) | listener, handler (a handler is server code or `ypy-on`) |
| out-of-band swap | an extra piece updated from the same response (`ypy-partial`, `ypy-swap-oob`) | side update, secondary swap |
| core | `ypy.min.js` | engine, runtime |
| extension | an optional script loaded after the core | plugin, addon, module |
| boost | `ypy-boost`: links and forms become requests | SPA mode |
| the older library | the library YPY came from; never named | any product name or old prefix |
| SDK | `github.com/ypy-develop/go` | server library, backend package |
