---
order: 100
title: Progress bar
description: Start a job with a button, poll its progress every second into a progress element, and let the server stop the polling with status 286.
---

# Progress bar

A button starts a job on the server. The page then asks for the job's progress every second and shows a bar. When
the job is done the server answers with the status `286`, the page swaps in the final markup and the polling stops.

## The page

The `ypy-poll` extension is what gives the status `286` its meaning.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-poll.min.js"></script>

<div id="job-area">
  <button ypy-post="/jobs" ypy-target="#job-area" ypy-swap="innerHTML">Start the job</button>
</div>
```

`POST /jobs` answers with the element that does the polling:

```html
<div id="job-7" ypy-get="/jobs/7" ypy-trigger="every 1s" role="status" aria-live="polite">
  <progress value="0" max="100">0%</progress> 0%
</div>
```

## The server (Go)

```go
package jobs

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	ypy "github.com/ypy-develop/go"
)

type Jobs struct {
	mu      sync.Mutex
	next    int
	started map[int]time.Time
}

func New() *Jobs { return &Jobs{started: map[int]time.Time{}} }

// percent is a stand-in for real work: one point every 100 ms.
func (j *Jobs) percent(id int) (int, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	t, ok := j.started[id]
	if !ok {
		return 0, false
	}
	return min(100, int(time.Since(t)/(100*time.Millisecond))), true
}

func view(id, pct int) string {
	return fmt.Sprintf(`<div id="job-%d" ypy-get="/jobs/%d" ypy-trigger="every 1s" role="status" aria-live="polite">`+
		`<progress value="%d" max="100">%d%%</progress> %d%%</div>`, id, id, pct, pct, pct)
}

func (j *Jobs) Start(w http.ResponseWriter, r *http.Request) {
	j.mu.Lock()
	j.next++
	id := j.next
	j.started[id] = time.Now()
	j.mu.Unlock()
	ypy.Fragment(w, view(id, 0))
}

func (j *Jobs) Status(w http.ResponseWriter, r *http.Request) {
	var id int
	if _, err := fmt.Sscan(r.PathValue("id"), &id); err != nil {
		http.NotFound(w, r)
		return
	}
	pct, ok := j.percent(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if pct < 100 {
		ypy.Fragment(w, `<progress value="`+fmt.Sprint(pct)+`" max="100">`+fmt.Sprint(pct)+`%</progress> `+fmt.Sprint(pct)+`%`)
		return
	}
	// 286 ends the polling of the element that asked; the body is swapped like any 2xx.
	ypy.FragmentStatus(w, 286, fmt.Sprintf(`Done. <a href="/jobs/%d/result">Download the result</a>`, id))
}

func Routes(j *Jobs) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", j.Start)
	mux.HandleFunc("GET /jobs/{id}", j.Status)
	return ypy.Middleware(mux)
}
```

The id is a number the server parsed, so it can go into the markup as it is. Anything that comes from a user goes
through `html.EscapeString` or `html/template`.

## How it works

- `ypy-post` on the button asks the server to start the job. `ypy-target="#job-area"` with `innerHTML` replaces the
  button with the element that polls.
- `ypy-trigger="every 1s"` makes the `div` ask `GET /jobs/7` once a second. It has no `ypy-target`, so the answer
  goes into the `div` itself, and the default `innerMorph` keeps the `<progress>` element and only changes its
  `value`.
- With `ypy-poll` loaded, a hidden tab polls four times slower. The server may also set the pace with a
  `YPY-Poll-Interval` header.
- The status `286` swaps its body like a `200` and then stops the polling, so the page needs no timer of its own.
  `ypy:poll:stop` fires on the element.

## Variations

- Without the extension: give the poller `ypy-swap="outerHTML"` and answer every poll with the whole element.
  The last answer leaves out `ypy-trigger`, so the element that replaces the poller has nothing to repeat. The
  status stays `200`.
- Slow down when it is busy. Call `w.Header().Set("YPY-Poll-Interval", "5s")` before `ypy.Fragment` in `Status`.
- A progress bar for a file upload needs no polling: see [file upload](file-upload.md).

## See also

- [`ypy-poll`](../extensions/poll.md) and [polling](polling.md)
- [Updating other content](updating-other-content.md) for a job that changes several parts of the page
