---
title: "Baseline: develop 47900c9e2"
description: "Starting-point numbers, recorded before the current results in bench/results/ (the raw file of that run is not kept). Times are script time in milliseconds (ms) or microseconds (us);"
order: 10
---
# Baseline: develop 47900c9e2

Starting-point numbers, recorded before the current results in `bench/results/` (the raw file of that run is not kept). Times are script time in milliseconds (ms) or microseconds (us); see plan.md for how to read them.

Build: the core of 47900c9e2 (byte-identical to dist/ypy.min.js at that commit)  sha 47900c9e2  2026-11-14  157.0.8103.51  Apple M4 Max  rounds 5

| metric | unit | median | p95 | min | noise % |
| --- | --- | ---: | ---: | ---: | ---: |
| swap.innerHTML.10 | ms | 0.416 | 0.498 | 0.348 | 2.3 |
| swap.outerHTML.10 | ms | 0.347 | 0.387 | 0.298 | 0.5 |
| swap.innerHTML.100 | ms | 3.38 | 4.22 | 2.73 | 2.6 |
| swap.outerHTML.100 | ms | 2.67 | 3.39 | 2.18 | 2.1 |
| swap.innerHTML.1000 | ms | 44.1 | 48.7 | 40.4 | 0.3 |
| swap.outerHTML.1000 | ms | 36.3 | 42.3 | 33.6 | 2.9 |
| swap.plain.innerHTML.100 | ms | 1.26 | 1.82 | 1.09 | 0.9 |
| swap.ids.innerHTML.100 | ms | 3.50 | 4.32 | 2.80 | 2.0 |
| swap.plain.innerHTML.1000 | ms | 24.2 | 29.3 | 22.3 | 2.4 |
| swap.ids.innerHTML.1000 | ms | 45.7 | 51.9 | 40.1 | 0.9 |
| morph.inner.10 | ms | 0.299 | 0.344 | 0.251 | 2.5 |
| morph.inner.10.preserve_failures | count | 0 | 0 | 0 | - |
| morph.outer.10 | ms | 0.247 | 0.297 | 0.215 | 0.4 |
| morph.outer.10.preserve_failures | count | 0 | 0 | 0 | - |
| morph.inner.100 | ms | 2.07 | 2.51 | 1.73 | 1.5 |
| morph.inner.100.preserve_failures | count | 0 | 0 | 0 | - |
| morph.outer.100 | ms | 1.40 | 1.74 | 1.22 | 3.8 |
| morph.outer.100.preserve_failures | count | 0 | 0 | 0 | - |
| morph.inner.1000 | ms | 21.6 | 24.0 | 19.3 | 0.5 |
| morph.inner.1000.preserve_failures | count | 0 | 0 | 0 | - |
| morph.outer.1000 | ms | 15.3 | 18.1 | 12.0 | 3.6 |
| morph.outer.1000.preserve_failures | count | 0 | 0 | 0 | - |
| morph.inner.noop.100 | ms | 1.84 | 2.41 | 1.64 | 0.8 |
| morph.inner.noop.100.preserve_failures | count | 0 | 0 | 0 | - |
| morph.inner.noop.1000 | ms | 18.1 | 21.5 | 16.3 | 0.4 |
| morph.inner.noop.1000.preserve_failures | count | 0 | 0 | 0 | - |
| process.rich | ms | 30.5 | 36.1 | 27.6 | 3.3 |
| process.plain | ms | 3.11 | 3.25 | 2.92 | 1.1 |
| attr.local | us | 0.122 | 0.175 | 0.105 | 4.1 |
| attr.inherit.hit.20 | us | 2.74 | 3.42 | 2.57 | 1.0 |
| attr.inherit.miss.20 | us | 2.44 | 2.55 | 2.30 | 0.3 |
| attr.inherit.hit.100 | us | 6.52 | 6.95 | 6.23 | 0.2 |
| attr.inherit.miss.100 | us | 6.24 | 6.36 | 5.97 | 0.8 |
| attr.inherit.append.20 | us | 2.44 | 2.66 | 2.26 | 2.8 |
| attr.inherit.implicit.100 | us | 6.64 | 6.77 | 6.32 | 0.7 |
| attr.mixed.20 | us | 2.72 | 2.90 | 2.60 | 0.7 |
| ctx.create.flat | us | 18.4 | 19.9 | 17.4 | 1.3 |
| ctx.create.deep20 | us | 31.4 | 32.5 | 30.0 | 0.2 |
| events.trigger.plain | us | 0.376 | 0.627 | 0.295 | 3.3 |
| events.trigger.native | us | 0.681 | 1.02 | 0.598 | 1.8 |
| events.trigger.native.listeners5 | us | 1.18 | 1.53 | 1.08 | 1.3 |
| events.on.click | us | 2.96 | 3.95 | 2.27 | 3.8 |
| events.filter.click | us | 12.8 | 13.8 | 11.7 | 2.2 |
| request.click | us | 74.4 | 87.5 | 64.8 | 0.9 |
| request.click.deep20 | us | 101 | 112 | 90.3 | 2.4 |
| request.overhead | us | 77.2 | 85.9 | 71.4 | 1.4 |
| request.overhead.listeners | us | 78.5 | 90.9 | 72.4 | 3.6 |
| request.events.empty.dispatched | count | 20.0 | 20.0 | 20.0 | - |
| request.events.empty.distinct | count | 10.0 | 10.0 | 10.0 | - |
| request.events.rows5.dispatched | count | 90.0 | 90.0 | 90.0 | - |
| request.events.rows5.distinct | count | 15.0 | 15.0 | 15.0 | - |
| session.total | ms | 545 | 561 | 533 | 3.4 |
| session.request | us | 1,015 | 2,710 | 65.0 | 4.4 |
| session.heap_peak | KB | 322 | 324 | 322 | 0.3 |
| session.heap_retained | KB | 284 | 286 | 284 | 0.4 |
| session.nodes_retained | count | 0 | 0 | 0 | - |
| session.listeners_retained | count | 0 | 0 | 0 | - |
| size.min | B | 39,884 | 39,884 | 39,884 | - |
| size.min_gzip | B | 14,260 | 14,260 | 14,260 | - |
| size.min_brotli | B | 12,813 | 12,813 | 12,813 | - |

note morph.inner.10: focus kept 560/560, value kept 560/560, selection kept 560/560

note morph.outer.10: focus kept 560/560, value kept 560/560, selection kept 560/560

note morph.inner.100: focus kept 180/180, value kept 180/180, selection kept 180/180

note morph.outer.100: focus kept 180/180, value kept 180/180, selection kept 180/180

note morph.inner.1000: focus kept 28/28, value kept 28/28, selection kept 28/28

note morph.outer.1000: focus kept 28/28, value kept 28/28, selection kept 28/28

note morph.inner.noop.100: focus kept 180/180, value kept 180/180, selection kept 180/180

note morph.inner.noop.1000: focus kept 28/28, value kept 28/28, selection kept 28/28

note request.events: empty response: 20 events dispatched for 10 distinct names; 5 wired rows: 90 for 15

note session.500: 500 requests, heap 1961 KB -> peak +323 KB -> end 2246 KB, nodes 1981 -> 1981, listeners 158 -> 158
