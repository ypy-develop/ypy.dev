---
title: "One request for identical GETs: the ypy-dedupe extension"
description: "Joins identical GET requests that are in flight at the same moment into one network request and gives every element the same answer."
order: 90
---
# One request for identical GETs: the `ypy-dedupe` extension

Two parts of a page often ask for the same fragment at the same moment: a poll and a click, several regions that load
the same menu, a user who double-clicks a link. Without help each one goes to the network and the server renders the
same thing twice. `ypy-dedupe` makes them share one request.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-dedupe.js"></script>

<button ypy-get="/cart" ypy-target="#cart">Refresh</button>
<div id="cart" ypy-get="/cart" ypy-trigger="every 5s"></div>
```

If the poll fires while the click's request is still in flight, only one `GET /cart` goes out. Both elements get the
answer and swap it into their own target.

The extension is not part of `ypy.max`: load it by itself.

## What counts as identical

Two requests are shared when all of these hold, and the first one has not answered yet:

- both are `GET`. A `POST`, `PUT`, `PATCH` or `DELETE` is never shared;
- the URL is the same, after it is resolved against the page (so `/cart` and `https://example.com/cart` on that origin
  match, and `/cart?a=1` and `/cart?a=2` do not);
- the fetch options `credentials`, `cache` and `mode` are the same;
- the request headers are the same, except the ones in `ypy.config.dedupe.ignore`.

The headers include the ones YPY adds, such as `YPY-Target` and `YPY-Request-Type`, and the ones set with
`ypy-headers`. They decide what the server renders, so two requests that differ in them are two requests. In the
example above both elements target `#cart`, so they match. If one of them targeted `#mini-cart`, the `YPY-Target` header
would differ and the requests would not be shared.

Once the shared request has answered, the next identical request starts a new one: this is not a cache. For a cache see
[Cache](cache.md).

## What each request gets

- Every request receives its own copy of the response, swaps it into its own target and fires its own events
  (`ypy:before:swap`, `ypy:after:swap`, and so on).
- An error is shared too: with a failed connection, or a `500`, every waiting request sees the error
  (`ypy:error` or `ypy:response:error` on each source element).
- A request that joined an existing one fires `ypy:dedupe:joined` on its source element. The one that started the
  network request does not.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-dedupe` | `false` | this element's GET requests are never shared |

The attribute is inherited with the usual `:inherited` form: `ypy-dedupe:inherited="false"` on a container keeps
everything inside it out of the sharing.

```html
<div ypy-dedupe:inherited="false">
  <button ypy-get="/report/refresh">Always ask the server</button>
</div>
```

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:dedupe:joined` | the source element of the request that joined | `ctx`, `url`, `waiters` (how many requests wait for the network request, this one included) |

```js
document.addEventListener('ypy:dedupe:joined', evt => {
  console.log(evt.detail.url, 'shared by', evt.detail.waiters);
});
```

The event is not cancelable.

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `dedupe.ignore` | `["YPY-Source"]` | names of request headers that do not make two requests different; compared without regard to case. Every other header must be equal |

`YPY-Source` says which element asked, and it differs for every element, so it is ignored by default. If you set
`dedupe.ignore` your list replaces the default: keep `YPY-Source` in it.

```html
<meta name="ypy-config" content='{"dedupe":{"ignore":["YPY-Source","YPY-Target","X-Trace"]}}'>
```

With `YPY-Target` ignored, two elements that target different containers share one request, and each answer is swapped
into its own target. Do this only when the server renders the same HTML whatever the target.

## Aborts

Each request keeps its own abort signal. The network request is aborted only when every request that waits for it
was aborted. If one element is replaced (and its request aborted) while another still waits, the other gets its answer.
A request made after that abort is a new request on the network.

## Together with `ypy-retry`

`ypy-dedupe` wraps the fetch of the request, as `ypy-retry` does. Load it before `ypy-retry` to retry each request on
its own; load it after to retry the shared request once for all of them.

## The server side

Nothing changes in the server: it sees one request and answers it once.

## Pitfalls

- Only requests in flight are shared. A second request after the first one answered goes to the network.
- Different `ypy-target` or `ypy-headers` means no sharing, as the headers differ. Check `ypy:dedupe:joined` to see
  that a join happens.
- Setting `dedupe.ignore` drops the default. Without `YPY-Source` in your list, elements never share a request.
- Do not ignore a header the server reads. If the server answers differently for a custom header set with
  `ypy-headers`, keep it out of `ignore`, or one element would get the answer made for another.
- Non-GET requests are never shared, by design: a repeated `POST` must reach the server.
