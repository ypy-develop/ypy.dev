---
title: "Primeiros passos"
description: "Em dez minutos você terá uma página em que um botão pede HTML ao servidor e a resposta troca um pedaço da página. Há dois caminhos de entrada, e os dois chegam ao mesmo lugar:"
order: 20
id: "GUIDE-getting-started-pt-BR"
type: "guide"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/guide", "lang/pt-BR"]
---

# Primeiros passos

Em dez minutos você terá uma página em que um botão pede HTML ao servidor e a resposta troca um pedaço da página. Há
dois caminhos de entrada, e os dois chegam ao mesmo lugar:

- [Com HTML puro](#caminho-a-html-puro): nenhum código de servidor, só uma pasta de arquivos estáticos.
- [Com Go ou Arandu](#caminho-b-go-e-arandu): um handler de verdade, que responde fragmentos e formulários.

Depois, [o que ler em seguida](#o-que-ler-em-seguida). Inglês: [Getting started](../../guides/getting-started.md).

## O que é o YPY

O YPY é uma biblioteca hipermídia. Um elemento HTML diz, em atributos, *qual requisição fazer*, *quando* e *onde colocar
a resposta*. O servidor responde com HTML, não com JSON. Não há etapa de build para você nem dependências: uma tag
`<script>`.

```html
<button ypy-get="/hello" ypy-target="#out"
        class="rounded bg-sky-600 px-4 py-2 text-white">Say hello</button>
<p id="out" class="mt-4 text-slate-700"></p>
```

Leia em voz alta: *ao clicar, faça GET em `/hello` e ponha a resposta em `#out`*. Sem o script o botão não faz nada,
e por isso a navegação de verdade (links e formulários) deve sempre funcionar sem ele; veja
[Melhoria progressiva](core-concepts.md#melhoria-progressiva).

## Pegue o arquivo

O build escreve em `dist/` os arquivos que você serve:

```sh
go run ./cmd/ypy build      # precisa de Go 1.26 ou mais novo; escreve dist/
```

| Arquivo | Use quando |
|---|---|
| `dist/ypy.min.js` | é um projeto novo |
| `dist/ypy.max.min.js` | é um projeto novo que quer as extensões oficiais (SSE, WebSocket, signals, preload e outras) |
| `dist/ext/*.min.js` | uma extensão por vez, depois do núcleo: veja [Extensões](../../extensions/using-and-writing.md) (em inglês) |
| `dist/ypy.d.ts` | tipos TypeScript de `window.ypy` |

Copie o arquivo de que precisa para junto das suas páginas. Os nomes de pacote nos registros ainda não foram
publicados, então não há `npm install` nem link de CDN para indicar aqui.

## Caminho A: HTML puro

Você precisa só de uma pasta e de qualquer servidor de arquivos estáticos. Dois arquivos:

`index.html`

```html
<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>YPY, primeira página</title>
  <script src="/ypy.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script> <!-- qualquer CSS serve; o Tailwind é só o exemplo -->
</head>
<body class="mx-auto max-w-md p-8 font-sans">
  <h1 class="text-2xl font-semibold">Olá, YPY</h1>

  <button ypy-get="/hello.html" ypy-target="#out"
          class="mt-4 rounded bg-sky-600 px-4 py-2 text-white">Dizer olá</button>

  <div id="out" class="mt-4"></div>
</body>
</html>
```

`hello.html` (um **fragmento**: sem `<html>`, sem `<head>`, só o pedaço):

```html
<p class="rounded border border-sky-200 bg-sky-50 p-3">Olá, vindo de um arquivo estático.</p>
```

Sirva a pasta e abra:

```sh
python3 -m http.server 8080     # ou qualquer servidor estático; depois abra http://localhost:8080
```

Clique no botão: o parágrafo aparece em `#out` e a barra de endereço não muda. Um arquivo estático não sabe
distinguir uma requisição do YPY de uma comum, então um link para `/hello.html` abriria o fragmento puro. Para um
primeiro teste serve, e é por isso que um servidor de verdade olha o cabeçalho `YPY-Request` (próximo caminho).

> A tag do Tailwind acima é o CDN de testes, bom para experimentar. Num projeto de verdade compile o seu CSS e
> acrescente as pastas de templates em `@source`; veja [Conceitos centrais](core-concepts.md#estilizar-os-estados).

## Caminho B: Go e Arandu

O SDK Go (`github.com/ypy-develop/go`) é o lado do servidor da mesma conversa. Precisa de Go 1.26 ou mais novo e de
nenhuma outra dependência.

```sh
go get github.com/ypy-develop/go
```

Um programa completo só com a biblioteca padrão. Ele serve a página, o script e um endpoint que responde de um jeito a
uma requisição do YPY e de outro a um navegador:

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"

	ypy "github.com/ypy-develop/go"
)

const page = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Items</title>
  <script src="/ypy.min.js"></script>
</head>
<body class="mx-auto max-w-md p-8">
  <form ypy-post="/items" ypy-target="#list" ypy-swap="beforeend" method="post" action="/items" class="flex flex-wrap gap-2">%s</form>
  <ul id="list" class="mt-4 space-y-1">%s</ul>
</body>
</html>`

// fields is the inside of the form, with an error when there is one. The page and the 422 answer share it.
func fields(value, problem string) string {
	invalid, message := "", ""
	if problem != "" {
		invalid = ` aria-invalid="true" aria-describedby="title-error"`
		message = `<p id="title-error" role="alert" class="w-full text-sm text-red-700">` + html.EscapeString(problem) + `</p>`
	}
	return `<input id="title" name="title" value="` + html.EscapeString(value) + `" required` + invalid +
		` class="flex-1 rounded border p-2 aria-invalid:border-red-500" aria-label="New item">` + message +
		`<button class="rounded bg-sky-600 px-4 py-2 text-white">Add</button>`
}

type app struct {
	mu    sync.Mutex
	items []string
}

func (a *app) list() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var sb strings.Builder
	for _, it := range a.items {
		sb.WriteString(row(it))
	}
	return sb.String()
}

func row(title string) string {
	return `<li class="rounded border p-2">` + html.EscapeString(title) + `</li>`
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, page, fields("", ""), a.list())
}

func (a *app) create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		ypy.ValidationError(w, fields(title, "Write something first.")) // the inside of the form, with its error
		return
	}
	a.mu.Lock()
	a.items = append(a.items, title)
	a.mu.Unlock()

	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/") // no script: post, redirect, get
		return
	}
	ypy.Fragment(w, row(title)) // goes into #list, at the end
}

func main() {
	a := &app{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("POST /items", a.create)
	mux.Handle("GET /ypy.min.js", http.FileServer(http.Dir("dist")))

	log.Fatal(http.ListenAndServe(":8080", ypy.Middleware(mux)))
}
```

Três pontos importam além deste exemplo:

1. **Um endpoint, duas respostas.** `ypy.FromRequest(r).IsYPY` é verdadeiro quando o cliente do YPY fez a requisição. Um
   navegador sem script (ou um robô de busca) recebe um redirecionamento e uma página inteira; o YPY recebe o fragmento.
2. **`422` para formulário inválido, `303` para o que deu certo.** O YPY põe o corpo do 422 *dentro do formulário* e
   segue o 303 como qualquer sucesso. Veja [Formulários](../../guides/forms.md) (em inglês).
3. **`ypy.Middleware` acrescenta `Vary: YPY-Request`.** Sem isso, um cache compartilhado pode entregar o fragmento a quem
   carrega a página.

O mesmo handler nas outras pilhas está nas [receitas](../recipes/go.md). Para ver o que o navegador envia e recebe, abra
as ferramentas do desenvolvedor: a requisição leva `YPY-Request: true`, `YPY-Target: ul#list` e
`YPY-Request-Type: partial`.

### Numa aplicação Arandu

Uma aplicação Arandu liga o mesmo handler onde a aplicação é montada, em geral `bootstrap/app.go`:

```go
router.Post("/items", items.Create, ypy.Middleware)
```

O roteador do Hesape recebe handlers e middleware comuns, então nada mais muda. CSRF, sessões e o layout vêm do Hesape;
veja a [receita do Arandu](../recipes/arandu.md).

## Um projeto escrito antes do YPY

Se você está levando um projeto que já tem marcação hipermídia, não comece aqui. Rode `ypy migrate`, que o converte
para o YPY e aponta a tag do script para o `ypy.min.js`: [Migração](migration.md).

## O que ler em seguida

| Você quer | Leia |
|---|---|
| entender o vocabulário | [Conceitos centrais](core-concepts.md) |
| levar um projeto escrito antes do YPY | [Migração](migration.md) |
| reagir ao que acontece na página | [Events](../../reference/events/index.md) (em inglês) |
| tratar formulários e validação | [Forms](../../guides/forms.md) (em inglês) |
| deixar seguro | [Checklist de segurança](security-checklist.md) |
| deixar rápido | [Performance](../../essays/performance.md) (em inglês) |
| testar o servidor sem navegador | [Testing your app](../../server/testing.md) (em inglês) |
| copiar um servidor pronto | [Receitas](../recipes/go.md): Go e Arandu; [Django](../../server/recipes/django.md), [Rails](../../server/recipes/rails.md) e [Laravel](../../server/recipes/laravel.md) em inglês |
| achar a página de qualquer tarefa | [Índice da documentação](../index.md) |
