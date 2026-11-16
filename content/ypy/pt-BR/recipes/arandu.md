---
title: "Receita: Arandu (Hesape e Kyse)"
description: "Como uma aplicação Go feita com o Arandu serve o YPY. O trabalho de rede é do SDK Go, o mesmo da receita Go;"
order: 20
id: "RECIPE-arandu-pt-BR"
type: "recipe"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/recipe", "lang/pt-BR", "stack/arandu"]
---

# Receita: Arandu (Hesape e Kyse)

Como uma aplicação Go feita com o Arandu serve o YPY. O trabalho de rede é do SDK Go, o mesmo da [receita
Go](go.md); o que o Arandu acrescenta é a ligação em `bootstrap/app.go`, o middleware de CSRF do Hesape, os helpers
dele de redirecionamento e de rejeição, e os componentes do Kyse. Inglês: [Recipe: Arandu](../../server/recipes/arandu.md).

O Hesape, o framework e o Kyse são os primeiros projetos levados para o YPY. A receita descreve o formato do código
convertido e usa os helpers que existem hoje no Hesape; um helper cujo nome ainda está mudando é descrito pelo que faz,
sem ser nomeado.

## Ligue onde a aplicação é montada

Uma aplicação Arandu liga tudo de forma explícita em `bootstrap/app.go`; não há contêiner nem descoberta automática. O
`ypy.Middleware` tem o formato padrão, então cabe na cadeia do roteador:

```go
// bootstrap/app.go
router.Get("/items", items.Index, ypy.Middleware)
router.Post("/items", items.Create, ypy.Middleware)
router.Get("/items/events", room.ServeHTTP, ypy.Middleware) // a stream: see the SDK README
```

O `ypy.Middleware` lê os cabeçalhos do YPY para o contexto da requisição (`ypy.FromContext(r.Context())`) e acrescenta
`Vary: YPY-Request` a toda resposta, então a página e o fragmento da mesma URL nunca são misturados por um cache.

Instale o SDK pelo repositório dele, como todo módulo do Arandu: `go get github.com/ypy-develop/go@<tag>`, com a tag que
você conferiu antes; não aponte o `go.mod` para uma cópia local com `replace`.

## O handler

Os handlers do Hesape são handlers `net/http` comuns. A parte do YPY é a mesma de qualquer lugar; a autorização e o
tenant são a parte do Arandu e ficam como em todo handler da sua aplicação (uma Policy por `security.Authorize`, o
tenant vindo de `data.Tenant(g)` do Grant que ela devolve, nunca do caminho, do corpo ou de um cabeçalho):

```go
type Items struct{ svc *itemsvc.Service }

func NewItems(svc *itemsvc.Service) *Items { return &Items{svc: svc} } // explicit wiring

func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	// 1. authorize: Policy -> security.Authorize -> Grant; the service takes the Grant, so the tenant is data.Tenant(g)
	// 2. validate and save (the service returns validation errors as values)
	item, errs, err := h.svc.Create(r.Context(), r.PostFormValue("title"))
	switch {
	case err != nil:
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	case len(errs) > 0:
		ypy.ValidationError(w, renderForm(r, errs)) // 422 and the form; see "Formulários inválidos" below
	case !ypy.FromRequest(r).IsYPY:
		ypy.SeeOther(w, "/items") // a browser without the script: post, redirect, get
	default:
		ypy.Fragment(w, renderRow(item)) // the new row, into #list
		ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, countText(r))
	}
}
```

`renderRow`, `renderForm` e `countText` são views do Hesape ou componentes do Kyse. Renderize o mesmo componente para a
página e para o fragmento, para existir um só lugar que desenha uma linha.

## Redirecionamentos

`hesape/http.Redirect(w, r, to)` já faz o que o `ypy.Redirect` faz: uma requisição do YPY recebe `YPY-Redirect` e
`204`, qualquer outro cliente recebe `303`. Use o que tiver à mão; não escreva o ramo à mão.

## CSRF

O middleware `CSRFProtect` do Hesape (pacote `hesape/http/middleware`) recusa uma requisição que muda estado e que o
navegador informa ser de outra origem (`403`), depois confere o token. Ele lê primeiro o cabeçalho `X-CSRF-Token` e
depois o campo de formulário `_token`, e responde `419` quando o token falta ou expirou.

- **Formulários** levam o campo oculto `_token`, então `<form ypy-post>` fica coberto com script ou sem script.
- **Requisições que nenhum formulário sustenta** (`ypy-delete` num botão, `ypy-patch` num interruptor) levam o token num
  cabeçalho. Ponha-o uma vez no layout. Um atributo só chega aos descendentes com `:inherited`:

```html
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}' class="bg-white text-slate-900">
```

- **Um token expirado** termina em `419`. O Hesape responde uma recusa a uma requisição do YPY com
  `YPY-Refresh: true`, então a página recarrega e mostra a recusa como uma página comum (o recarregamento é um `GET`,
  que nunca é conferido, e traz um token novo). Nada entra em laço.

## Formulários inválidos: dois jeitos, os dois funcionam

1. **Flash e redirecionamento (o que o `Reject` do Hesape faz).** A resposta a um formulário rejeitado é um
   redirecionamento de volta, levando as mensagens e o que foi digitado no flash. O YPY segue o redirecionamento
   (`YPY-Redirect` numa requisição do YPY, `303` nas outras) e mostra a página que volta. Não exige mudança ao ir para o
   YPY e funciona também para um navegador sem script. O custo é uma segunda requisição.
2. **`422` no lugar.** `ypy.ValidationError(w, renderForm(r, errs))` responde o formulário com os erros; o YPY o funde
   sobre o formulário, mantém o que foi digitado e foca o primeiro campo inválido. Acrescente
   `ypy-error-swap="outerMorph"` no formulário quando a resposta é o `<form>` inteiro. Uma requisição, sem flash. Veja
   [Forms](../../guides/forms.md) (em inglês).

Escolha um por aplicação e mantenha. Não misture os dois num formulário.

## O layout e o Tailwind

O layout carrega o script, o cabeçalho de CSRF e a folha de estilo. Sirva o `ypy.min.js` pelo caminho de assets da aplicação:

```html
<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="{{ asset "app.css" }}">
  <script src="{{ asset "ypy.min.js" }}"></script>
</head>
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}'>
  <main id="main" class="mx-auto max-w-3xl p-6">{{ template "content" . }}</main>
</body>
</html>
```

`asset` representa como a sua aplicação resolve a URL de um asset. Para o Tailwind, escaneie os templates do Kyse e do
Go e acrescente as variantes de estado uma vez ([Estilizar os estados](../guides/core-concepts.md#estilizar-os-estados)):

```css
@import "tailwindcss";
@source "../internal/**/*.go";
@source "../views/**/*";
@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

## Kyse

Os componentes do Kyse que iniciam requisições renderizam os atributos do YPY. Use as props do próprio componente para o
verbo, o alvo e a troca, em vez de escrever o atributo numa string, para o componente manter o escape e a acessibilidade
com que foi construído.

Uma aplicação escrita antes do YPY tem fontes `.kyse.go` e os arquivos `.go` compilados ao lado. O `ypy migrate` converte
os dois do mesmo jeito, e o `aru view:build` regenera os compilados a partir das fontes com o mesmo resultado; veja
[Migration](../../migration/migrate.md#step-by-step-for-an-arandu-or-kyse-project) (em inglês) para a ordem (hesape,
framework, kyse, depois os produtos).

## Testes

Teste o handler sem navegador, com o `ypytest` no `go test` da aplicação:

```go
res := ypytest.Do(handler, ypytest.NewRequest(http.MethodPost, "/items",
	ypytest.Target("ul#list"),
	ypytest.Header("X-CSRF-Token", token),
	ypytest.Form(url.Values{"title": {"Buy milk"}})))
res.AssertStatus(t, http.StatusOK)
res.AssertVary(t, "YPY-Request")
```

[Testing your app](../../server/testing.md) (em inglês) tem o resto.

## Checklist

- [ ] `ypy.Middleware` na cadeia de toda rota que responde um fragmento e uma página
- [ ] a Policy e o Grant decidem, o tenant vem de `data.Tenant(g)`, a marcação é o mesmo componente na página e no fragmento
- [ ] o `CSRFProtect` está ligado, o layout leva o cabeçalho, o `419` recarrega a página
- [ ] um só jeito de responder formulário inválido, redirecionamento com flash ou `422`
- [ ] o registro de atividade grava a ação como em qualquer outra requisição: o YPY não muda isso
- [ ] `ypy check` e `ypy migrate` rodaram nas fontes `.kyse.go` e nos arquivos compilados
