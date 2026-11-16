---
title: "Receita: Go (net/http)"
description: "Um servidor completo com a biblioteca padrão e o SDK Go do YPY: uma página, um formulário que acrescenta um item, um fragmento para o YPY, um 303 para o navegador sem script, um 422 para formulário..."
order: 20
id: "RECIPE-go-pt-BR"
type: "recipe"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/recipe", "lang/pt-BR", "stack/go"]
---

# Receita: Go (`net/http`)

Um servidor completo com a biblioteca padrão e o SDK Go do YPY: uma página, um formulário que acrescenta um item, um
fragmento para o YPY, um `303` para o navegador sem script, um `422` para formulário inválido, `Vary: YPY-Request`,
proteção contra CSRF e classes do Tailwind. Inglês: [Recipe: Go](../../server/recipes/go.md). O teste dele está em
[Testing your app](../../server/testing.md) (em inglês).

Outras pilhas: [Arandu](arandu.md) · [Django](../../server/recipes/django.md) · [Rails](../../server/recipes/rails.md) ·
[Laravel](../../server/recipes/laravel.md) (em inglês).

## O que o servidor precisa fazer

| O servidor | Como |
|---|---|
| distingue uma requisição do YPY de uma carga de página | `ypy.FromRequest(r).IsYPY`, que lê `YPY-Request` |
| responde um fragmento ao YPY e uma página a todo o resto | um handler, dois ramos |
| acrescenta `Vary: YPY-Request` | o `ypy.Middleware` faz isso em toda resposta |
| responde formulário inválido com `422` e o formulário | `422`, o corpo é o formulário com os erros; com `ypy-error-swap="outerMorph"` no formulário a resposta é o `<form>` inteiro |
| responde formulário válido com `303` | `ypy.SeeOther`; o cliente segue e troca a página final |
| recusa requisições de outros sites | `http.CrossOriginProtection` (Go 1.25 em diante) |
| atualiza mais de um lugar | `ypy.Partial` uma vez por alvo, depois do fragmento principal |

## O programa

```sh
go get github.com/ypy-develop/go    # Go 1.26 ou mais novo
```

`main.go`:

```go
package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	ypy "github.com/ypy-develop/go"
)

var tpl = template.Must(template.New("").Parse(`
{{define "layout"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Items</title>
  <link rel="stylesheet" href="/app.css">
  <script src="/ypy.min.js"></script>
</head>
<body class="mx-auto max-w-lg p-6">
  <h1 class="text-2xl font-semibold">Items <span id="count" class="text-sm text-slate-500">{{len .Items}} items</span></h1>
  {{template "form" .}}
  <ul id="list" class="mt-4 space-y-1">{{range .Items}}{{template "row" .}}{{end}}</ul>
</body>
</html>{{end}}

{{define "form"}}<form id="new" ypy-post="/items" ypy-target="#list" ypy-swap="beforeend" ypy-error-swap="outerMorph"
      method="post" action="/items" class="flex flex-wrap gap-2 ypy-error:ring-1 ypy-error:ring-red-500">
  <input id="title" name="title" value="{{.Title}}" required
         {{if .Error}}aria-invalid="true" aria-describedby="title-error"{{end}}
         class="flex-1 rounded border p-2 aria-invalid:border-red-500" aria-label="New item">
  <button class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50">Add</button>
  {{if .Error}}<p id="title-error" role="alert" class="w-full text-sm text-red-700">{{.Error}}</p>{{end}}
</form>{{end}}

{{define "row"}}<li class="rounded border p-2">{{.}}</li>{{end}}
{{define "count"}}{{.}} items{{end}}
`))

type view struct {
	Items []string
	Title string
	Error string
}

type app struct {
	mu    sync.Mutex
	items []string
}

func (a *app) snapshot() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.items...)
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.ExecuteTemplate(w, "layout", view{Items: a.snapshot()})
}

func (a *app) create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		tpl.ExecuteTemplate(w, "form", view{Title: title, Error: "Write something first."})
		return
	}
	a.mu.Lock()
	a.items = append(a.items, title)
	n := len(a.items)
	a.mu.Unlock()

	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/")
		return
	}
	var row, count strings.Builder
	tpl.ExecuteTemplate(&row, "row", title)
	tpl.ExecuteTemplate(&count, "count", n)
	ypy.Fragment(w, row.String())
	ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, count.String())
}

func main() {
	a := &app{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("POST /items", a.create)
	mux.Handle("GET /ypy.min.js", http.FileServer(http.Dir("dist")))

	csrf := http.NewCrossOriginProtection()
	log.Fatal(http.ListenAndServe(":8081", csrf.Handler(ypy.Middleware(mux))))
}
```

Ponha o `ypy.min.js` (de `dist/`) numa pasta `dist` ao lado do binário e rode:

```sh
go run .          # depois abra http://localhost:8081
```

## Como funciona

**A página.** `index` renderiza o layout com a lista inteira. Um link ou um recarregamento sempre recebe isto, com ou
sem script: é a melhoria progressiva.

**O formulário.** `ypy-post="/items"` envia; `ypy-target="#list" ypy-swap="beforeend"` acrescenta a resposta ao fim da
lista; `method` e `action` o mantêm funcionando sem script. Escrever `ypy-error-swap="outerMorph"` diz ao YPY que,
quando o servidor responde `422`, a resposta é o `<form>` inteiro, fundido sobre o atual, então o que o usuário digitou
permanece e o foco vai para o primeiro campo inválido ([Forms](../../guides/forms.md), em inglês).

**O caso válido.** `create` responde a nova linha (`ypy.Fragment`) e, na mesma resposta, um `<ypy-partial>` que atualiza
o contador do título ([Multiple targets](../../guides/multiple-targets.md), em inglês). Um navegador sem script recebe um
`303` para `/` e recarrega a página com o item.

**O `422`.** O template do formulário é renderizado de novo com o erro; `aria-invalid` e `aria-describedby` ligam a
mensagem ao campo, `role="alert"` faz o leitor de tela falar.

**`Vary`.** O `ypy.Middleware` acrescenta `Vary: YPY-Request`. A mesma URL (`/`) pode responder uma página ou um
fragmento conforme esse cabeçalho, e um cache não pode misturar os dois.

## CSRF

`http.NewCrossOriginProtection()` (biblioteca padrão, Go 1.25 em diante) rejeita com `403` uma requisição que muda
estado e que o navegador informa ser de outro site, lendo `Sec-Fetch-Site` e, na falta, `Origin`. Deixa passar
requisições da mesma origem e de clientes que não enviam nenhum dos dois cabeçalhos (ferramentas de linha de comando),
então é uma boa primeira camada e não custa marcação. Envolva todo o mux, como no programa acima.

Se você quer também um token (uma API usada por outras origens, defesa em profundidade, ou um auditor que pede), envie-o
num cabeçalho e confira num middleware:

```go
func verifyToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get("X-CSRF-Token")
		if got == "" {
			got = r.PostFormValue("_token") // a plain form post carries it in a hidden field
		}
		want := tokenFor(r) // yours: derived from the session, never a constant
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

e ponha o token no layout, uma vez, para toda requisição do YPY levá-lo:

```html
<meta name="csrf-token" content="{{ .CSRFToken }}">
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}'>
```

(o `html/template` escapa o valor para o contexto em que está.) Um formulário sem script acrescenta
`<input type="hidden" name="_token" value="{{ .CSRFToken }}">`.

## Tailwind

As classes dos templates precisam ser escaneadas pelo build, ou uma classe que só aparece num fragmento é removida do
CSS. Com o Tailwind v4, em `app.css`:

```css
@import "tailwindcss";
@source "./**/*.go";          /* the templates live in Go strings here; point to .html or .tmpl files if you have them */

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

As duas variantes fazem funcionar `ypy-request:opacity-50` e `ypy-error:ring-red-500` no template:
[Estilizar os estados](../guides/core-concepts.md#estilizar-os-estados).

## Outras coisas que um handler Go faz

```go
ypy.Redirect(w, r, "/login")                     // YPY: YPY-Redirect + 204; browser: 303
ypy.Location(w, ypy.LocationOptions{Path: "/items/7", Target: "#main"}) // navigate without a full load
ypy.PushURL(w, "/items?page=2")                  // change the address bar
ypy.Retarget(w, "#errors"); ypy.Reswap(w, "innerHTML") // change where and how this answer goes
ypy.Trigger(w, "item:saved", map[string]any{"count": n}) // fire a client event
ypy.Refresh(w)                                   // reload the page
```

Uma restauração de histórico (o botão voltar) chega **sem** `YPY-Request` e com `YPY-History-Restore-Request: true`:
responda uma página inteira. `ypy.FromRequest(r).HistoryRestore` informa. O streaming (`ypy.NewStream`,
`ypy.Broadcaster`) está no README do SDK.

## Checklist

- [ ] `ypy.Middleware` envolve as rotas que podem responder fragmento e página
- [ ] `422` para formulário inválido, com o formulário e `ypy-error-swap="outerMorph"` (ou só os campos)
- [ ] `303` depois de um `POST` que deu certo, para o navegador sem script
- [ ] todo valor num template é escapado pelo `html/template`; nenhum texto do usuário num nome de atributo
- [ ] a proteção contra CSRF envolve o mux; requisições sem formulário levam o cabeçalho do token
- [ ] um teste por público: YPY, navegador comum, formulário inválido ([Testing](../../server/testing.md), em inglês)
