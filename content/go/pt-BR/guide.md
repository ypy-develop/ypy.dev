---
title: "SDK Go: streaming, signals, cabeçalhos, CSRF, robôs de busca e testes"
description: "As partes do SDK além de um handler simples. O exemplo de handler e a instalação estão no README do repositório; os helpers de requisição e resposta estão no índice da documentação."
order: 20
---
# SDK Go: streaming, signals, cabeçalhos, CSRF, robôs de busca e testes

As partes do SDK além de um handler simples. O exemplo de handler e a instalação
estão no README do repositório; os [helpers de requisição e resposta](README.md) estão
no índice da documentação. O formato do streaming está em
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

## Streaming

```go
func (h *Jobs) Events(w http.ResponseWriter, r *http.Request) {
	s, err := ypy.NewStream(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer s.Close() // sempre: isso para o keepalive

	s.PatchElements(`<p id="status">Trabalhando</p>`)       // morph por id
	s.PatchSignals(map[string]any{"progress": 50})          // JSON Merge Patch
	s.PatchElements(`<li>pronto</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#log"))
	s.Event("job:done", map[string]int{"id": 7})
	s.Redirect("/jobs/7")
	<-s.Done() // volta quando o cliente sai
}
```

Para mandar os mesmos eventos a muitos clientes, monte um `Broadcaster` como handler
da rota do stream e publique de qualquer lugar:

```go
room := ypy.NewBroadcaster(ypy.WithSnapshot(func(r *http.Request) ([]ypy.Event, error) {
	ev, err := ypy.PatchElementsEvent(renderMessages()) // o estado completo de agora
	return []ypy.Event{ev}, err
}))
defer room.Close()

mux.Handle("GET /room/events", room)

room.PatchElements(`<li>olá</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"))
```

O snapshot é o que torna a reconexão segura: toda conexão começa pelo estado
completo, e o cliente faz o morph dele. Um cliente que não acompanha é desconectado e
reconecta sozinho. Nada bloqueia e nenhuma goroutine fica para trás.

### Listas: `WithKey` e `WithLimit`

Um chat, um log ou um feed acrescenta itens numa ponta da lista. Duas opções num patch
`append` ou `prepend` mantêm essa lista correta, sem extensão na página:

```go
room.PatchElements(`<li data-id="42">Ana: olá</li>`,
	ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"),
	ypy.WithKey("data-id"), ypy.WithLimit(100))
```

- `WithKey(attr)`: um item cujo atributo tem o valor de um item que já está na lista
  o substitui no mesmo lugar; senão é acrescentado. O mesmo valor duas vezes no
  mesmo patch fica com o último. Um item sem o atributo é sempre acrescentado.
- `WithLimit(n)`: depois do patch a lista tem no máximo `n` itens; os mais antigos
  saem (os primeiros num `append`, os últimos num `prepend`).
- A chave vale antes do limite, então atualizar um item que já aparece nunca empurra
  outro para fora. Todo elemento filho do alvo conta como item: deixe cabeçalho e
  mensagem de lista vazia fora da lista.

São as regras dos modificadores `key:` e `limit:` do
[`ypy-list`](../../ypy/extensions/list.md); o cliente do stream não precisa dessa
extensão. `PatchElementsEvent` devolve erro para `WithKey` ou `WithLimit` em qualquer
modo que não seja `append` ou `prepend`, para limite menor que 1 e para uma chave que
não é nome de atributo (letras, dígitos, `_`, `.` e `-`). Sem seletor, o patch é o
próprio elemento da lista (casado por `id`) com os itens novos dentro. O snapshot do
`Broadcaster` continua valendo: ao reconectar, o cliente faz o morph da lista
inteira, e os patches seguintes continuam a partir dela.

### Refresh: mandar todos os clientes buscarem de novo

Um patch leva o mesmo HTML para todo mundo. Quando a mudança aparece diferente para
cada usuário (um carrinho, um saldo, uma lista filtrada por permissão, uma página com
token CSRF), mande `ypy-refresh`. Ele não leva HTML: cada cliente busca a sua própria
cópia, com a sua própria sessão, e a resposta entra por morph.

```html
<div id="room" ypy-stream="/room/events" ypy-history-elt>
  <section id="cart" ypy-get="/cart">...</section>
</div>
```

```go
// depois que um pedido muda: cada cliente conectado busca o seu #cart
room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))

room.Refresh() // sem seletor: cada cliente busca a sua página de novo
```

`Stream.Refresh` faz o mesmo num stream só, e `RefreshEvent` monta o valor para
`Send` e `Publish`. Opções:

| Opção | Significado |
|---|---|
| `WithRefreshSelector(sel)` | os elementos a buscar de novo; cada um que casar precisa ter `ypy-get`. Sem ela, a página |
| `WithRefreshDebounce(d)` | eventos com o mesmo seletor dentro de `d` do primeiro viram um pedido só. Padrão 250 ms (`DefaultRefreshDebounce`), de 0 a 1 minuto |
| `WithRefreshJitter(d)` | cada cliente espera um tempo aleatório de até `d` depois do debounce, para mil clientes não chegarem juntos. De 0 a 1 minuto |

Um valor fora de 0 a 1 minuto faz o `RefreshEvent` devolver erro. O SDK sempre
escreve a linha `debounce`, porque um quadro SSE sem linha `data` não é entregue.

O que o cliente pede é uma requisição YPY normal (mesmos cabeçalhos, cookies,
signals):

- **Com seletor**, um `GET` para a URL do `ypy-get` de cada elemento que casa, com o
  resultado trocado nos filhos desse elemento. O handler responde um fragmento, como
  para qualquer `ypy-get`. Um elemento sem `ypy-get` é reportado por `ypy:stream:error`.
- **Sem seletor**, um `GET` para o caminho e a query da página, com
  `YPY-Request-Type: full`. `q.IsPartial()` é falso, então o mesmo handler que
  responde ao navegador responde a página, e o cliente faz o morph do elemento
  marcado com `ypy-history-elt` (ou do body) a partir dela. Não responda um fragmento
  pelado ali.

O cliente dispara o evento cancelável `ypy:stream:refresh` no elemento que abriu o
stream, logo antes do pedido. Não use refresh para uma mudança igual para todos: um
patch custa menos. O refresh alcança os clientes conectados quando a mudança
acontece; quem conecta depois recebe o snapshot. O lado do cliente está em
[Streaming](../../ypy/extensions/stream.md).

## Signals

```go
type Search struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
}

s, err := ypy.ReadSignals[Search](r) // ErrNoSignals quando o pedido não traz nenhum
ypy.PatchSignals(w, map[string]any{"page": 1, "draft": nil}) // null apaga
```

O cliente manda os signals no parâmetro `ypy-signals` da URL (GET), como corpo JSON
(POST, PUT, PATCH) ou no campo de formulário `ypy-signals`.

### Signals assinados, somente leitura

Um signal declarado com `ypy-signals.server` não pode ser escrito pela página, mas o
navegador devolve todos os signals, então um usuário pode mandar qualquer valor. Para
confiar num valor, assine-o ao enviar e confira ao receber (HMAC-SHA256 do nome e do
JSON canônico do valor, comparação em tempo constante):

```go
signer, _ := ypy.NewSigner(newKey, oldKey) // assina com newKey; aceita as duas enquanto houver páginas com valores antigos
signer = signer.Bind(sessionID)             // opcional: um valor assinado para uma sessão falha em outra

// Enviando: os valores e as assinaturas, para a página ou para um stream.
decl, _ := signer.Declaration(map[string]any{"plan": "free", "seats": 3}) // <div ypy-signals.server="{{decl}}">
ypy.PatchSignedSignals(w, signer, map[string]any{"plan": "pro"})          // ou: patch, _ := signer.SignPatch(...); stream.PatchSignals(patch)

// Recebendo: a conferência acontece antes de qualquer decodificação.
got, err := ypy.ReadSignedSignals[Account](r, signer, "plan", "seats")
if errors.Is(err, ypy.ErrSignatureMissing) || errors.Is(err, ypy.ErrSignatureInvalid) {
	http.Error(w, "forbidden", http.StatusForbidden) // errors.As com *ypy.SignatureError dá o nome do signal
	return
}
```

As assinaturas viajam num signal reservado, `ypySig`; o cliente não muda. Nomeie os
signals em que você confia: um signal ausente ou sem assinatura falha. O formato está
na seção 3.3 do [`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).
Nos testes, `ypytest.SignedSignals(signer, values)` envia um pedido assinado.

## Cabeçalhos das extensões

As extensões `ypy-validate`, `ypy-layers` e `ypy-cache` conversam por cabeçalhos:

```go
q := ypy.FromRequest(r)
if q.IsValidation() { // YPY-Validate: o nome do campo conferido; não altere nada
	if msg := problems(r)[q.Validate]; msg != "" {
		ypy.ValidationError(w, "<span>"+html.EscapeString(msg)+"</span>")
		return
	}
	ypy.NoContent(w)
	return
}
if q.InLayer() { /* YPY-Layer: "modal", "drawer" ou "popup": responda só o conteúdo */ }

ypy.OpenLayer(w, "/items/7/confirm", ypy.LayerOptions{Mode: ypy.LayerDrawer, CloseOn: []string{"deleted"}}) // YPY-Open-Layer
ypy.ExpireCache(w, "/reports*", "/dashboard")                                                              // YPY-Cache-Expire
```

`OpenLayer` escreve só a URL quando não há opções e JSON (ASCII) nos outros casos;
`OpenLayerHTML` mostra um markup que você já tem. `KeepCache` manda `false`. Nos
testes: `ypytest.Validate("email")`, `ypytest.Layer(ypy.LayerModal)`,
`res.AssertOpenLayer`, `res.AssertCacheExpire`.

## Proteção CSRF

`ypy.NewCSRF` protege toda requisição não segura (tudo menos GET, HEAD, OPTIONS e
TRACE) com um token preso à sessão: o HMAC-SHA256 do id da sessão, então não guarda
nada. O lado Go e a extensão [`ypy-csrf`](../../ypy/extensions/csrf.md) da página são
as duas metades de um mesmo contrato:

1. O servidor escreve o token da sessão no head da página:
   `<meta name="ypy-csrf" content="TOKEN">` (`csrf.Meta(r)`).
2. A extensão (`ext/ypy-csrf.js`, listada em `ypy.config.extensions` quando você a
   define) manda esse token em `X-CSRF-Token` em todo `POST`, `PUT`, `PATCH`,
   `DELETE` e `QUERY` para a origem da própria página, mesmo que o elemento não herde
   cabeçalho nenhum. Nunca manda para outra origem. Um formulário comum, sem cliente,
   manda o token no campo oculto `csrf_token` (`csrf.Field(r)`).
3. O `csrf.Middleware` lê o cabeçalho, ou o campo, e compara com o token da sessão da
   requisição. Também confere o `Origin`.
4. Quando o token precisa mudar, o servidor responde com `YPY-CSRF: <token novo>` e a
   extensão troca o conteúdo do meta. O próximo pedido já leva o novo.

```go
csrf, err := ypy.NewCSRF(key, // 32 bytes aleatórios ou mais
	ypy.CSRFSession(func(r *http.Request) string { return sessions.ID(r) }), // opcional, veja abaixo
	ypy.CSRFPreviousKeys(oldKey))                                              // opcional, para rotação
handler := ypy.Middleware(csrf.Middleware(mux)) // 403 com corpo claro quando o token falta ou não confere
```

```html
<head>{{ .CSRFMeta }}</head>                <!-- csrf.Meta(r): <meta name="ypy-csrf" content="..."> -->
<form method="post">{{ .CSRFField }}</form> <!-- csrf.Field(r): um formulário comum, sem cliente -->
```

Ponha `csrf.Meta(r)` (e `csrf.Field(r)` nos formulários que precisam funcionar sem o
cliente) nos dados da página. Quem não tem sessão recebe uma aleatória num cookie
(`HttpOnly`, `SameSite=Lax`, prefixo `__Host-` sob TLS ou com `ypy.CSRFSecure()`),
então um formulário de login também fica protegido; `CSRFSession` dá o id da sessão
da própria aplicação, que vale quando não é vazio. O middleware também acrescenta
`Vary: Cookie`, já que a página depende do cookie.

O middleware responde `403` com um corpo em texto que diz o que fazer:

| Erro | Causa | Resposta |
|---|---|---|
| `ErrCSRFOrigin` | o `Origin` é de outra origem (e não está em `CSRFTrustedOrigins`), ou não há `Origin` e vem `Sec-Fetch-Site: cross-site` | 403 |
| `ErrCSRFMissing` | nenhum token no cabeçalho nem no campo. Com o `ypy-csrf` carregado, a página não tem o meta: o cliente dispara `ypy:csrf:missing` | 403 |
| `ErrCSRFInvalid` | o token é de outra sessão (um login em outra aba), ou o cookie da sessão não chegou | 403 com `YPY-CSRF` trazendo o token atual, e o cliente tenta de novo com ele |
| válido, feito com uma chave anterior | rotação de chave | o pedido segue, e a resposta leva `YPY-CSRF` |

`CSRFOnFailure(fn)` troca o corpo do 403. `csrf.Check(r)` é a mesma checagem como
erro, para um handler fora do middleware. Atrás de um proxy que reescreve o `Host`,
liste a origem pública em `CSRFTrustedOrigins`.

Quando a sessão muda, a página ainda guarda o token antigo. Chame
`csrf.RenewFor(w, novoIDDeSessao)` (ou `csrf.Renew(w, r)`) no handler de login:

```go
mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
	id := sessions.Start(w, r.PostFormValue("email")) // a sessão nova
	csrf.RenewFor(w, id)                               // YPY-CSRF: o cliente atualiza o meta
	ypy.Redirect(w, r, "/")
})
```

Um stream aberto com GET não pode mudar estado (por isso não precisa de token); uma
ação disparada a partir de um stream é um POST como outro qualquer. Isto não serve
para APIs autenticadas por token em cabeçalho (sem cookie, não há CSRF). O formato
está na seção 3.4 do
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

Teste com `ypytest.CSRFToken`, pegando o token de uma sessão com `TokenFor`:

```go
func TestCreateNeedsTheToken(t *testing.T) {
	csrf, _ := ypy.NewCSRF(key, ypy.CSRFSession(func(r *http.Request) string { return "user-7" }))
	h := csrf.Middleware(mux)
	form := ypytest.Form(url.Values{"title": {"Comprar leite"}})

	ypytest.Do(h, ypytest.NewRequest("POST", "/items", form)).AssertStatus(t, http.StatusForbidden)

	token := csrf.TokenFor("user-7")
	ypytest.Do(h, ypytest.NewRequest("POST", "/items", form, ypytest.CSRFToken(token))).
		AssertStatus(t, http.StatusOK)

	stale := ypytest.Do(h, ypytest.NewRequest("POST", "/items", form, ypytest.CSRFToken(csrf.TokenFor("user-8"))))
	stale.AssertStatus(t, http.StatusForbidden)
	stale.AssertHeader(t, ypy.HeaderCSRF, token) // o cliente pega o token certo
}
```

## Fragmentos e robôs de busca

Um fragmento é um pedaço de página. Um buscador ou um robô de IA que chegue a um (por
um cache que ignora o `Vary`, ou por um link vazado) não deve colocá-lo num índice. O
SDK marca os fragmentos por você:

| Onde | O que acrescenta |
|---|---|
| `ypy.Fragment`, `FragmentStatus`, `ValidationError`, `Partial` | `X-Robots-Tag: noindex` e `Vary: YPY-Request` |
| `ypy.Middleware` | `Vary: YPY-Request` em toda resposta, e `X-Robots-Tag: noindex` na resposta de uma requisição YPY que espera um fragmento (`q.IsPartial()`) |

Nenhum dos dois mexe numa página. O cabeçalho fica de fora quando o markup passado ao
`Fragment` é um documento inteiro (doctype ou elemento `<html>`), para que um handler
que escreve a página completa assim não suma dos resultados de busca; um pedido do
navegador, uma restauração de histórico e um pedido de conteúdo completo nunca o
recebem do middleware; e um valor que o handler já pôs em `X-Robots-Tag` é mantido.

Um fragmento que é uma página por si só, que um robô deve indexar, desliga isso. O
`Vary` continua, porque ele impede o cache de misturar fragmento e página:

```go
ypy.Fragment(w, docsListHTML, ypy.AllowIndex())

handler := ypy.NewMiddleware(ypy.AllowIndex())(mux) // no lugar de ypy.Middleware
```

O cabeçalho não é do YPY e o cliente o ignora. Num teste, `res.AssertNoIndex(t)`
confere que a resposta tem `noindex` (ou `none`) que não vale só para um robô
(`googlebot: noindex` não conta) e `YPY-Request` no `Vary`.

## Testes sem navegador

```go
func TestCreate(t *testing.T) {
	res := ypytest.Do(handler, ypytest.NewRequest("POST", "/items",
		ypytest.Target("ul#items"),
		ypytest.Form(url.Values{"title": {"Comprar leite"}})))

	res.AssertStatus(t, http.StatusOK)
	res.AssertText(t, "li", "Comprar leite")               // seletor CSS
	res.AssertPartial(t, "#count", "1 itens")              // <ypy-partial>
	res.AssertTrigger(t, ypy.HeaderTrigger, "item:saved", map[string]any{"count": 1})
	res.AssertVary(t, ypy.HeaderRequest)
}
```

O `ypytest` também lê patches de signals (`res.AssertSignals`), os eventos de um
stream (`res.AssertPatchElements`, `res.AssertRefreshEvent`, `res.Events()`) e roda
handlers que nunca terminam (`ypytest.OpenStream`). O parser de HTML é pequeno e feito
para o markup que um handler escreve; usa só a biblioteca padrão.

### Fragmento ou página: qual helper

Um handler que atende uma URL que o cliente empurra no histórico precisa responder
duas formas. Uma requisição YPY quer um fragmento, sem layout. Uma requisição do
navegador (um reload, um link compartilhado) e a restauração de histórico depois de
voltar ou avançar querem a página inteira. O defeito clássico é uma resposta só para
os dois: o usuário recarrega `/items` e vê uma lista pelada. Os testes para isso:

| Helper | O que confere | Use quando |
|---|---|---|
| `res.AssertFragment(t)` | sem doctype e sem elemento `html`, `head` ou `body`; corpo vazio (um 204) passa | você testa uma resposta a um pedido YPY (um POST de formulário, um partial) |
| `res.AssertFullPage(t)` | um elemento `html` e um `body` | você testa a resposta a um pedido do navegador ou a um pedido `Full()` |
| `ypytest.AssertFragmentAndPage(t, h, url, opts...)` | três `GET` na mesma URL (YPY com `opts`, comum com `NotYPY()`, `HistoryRestore()`): fragmento, página, página, e `YPY-Request` no `Vary` das três. Devolve o fragmento e a página | a URL é uma página que também serve fragmento (rotas com `ypy-push-url`, links boosted, `ypy-history-elt`) |
| `res.AssertNoIndex(t)` | `X-Robots-Tag: noindex` e `Vary: YPY-Request` | a resposta é um fragmento que deve ficar fora dos buscadores |

```go
func TestItemsPage(t *testing.T) {
	fragment, page := ypytest.AssertFragmentAndPage(t, handler, "/items", ypytest.Target("#items"))
	fragment.AssertSelector(t, "li", 3)
	fragment.AssertNoIndex(t)
	page.AssertText(t, "title", "Itens")
}
```

O `AssertFragmentAndPage` só faz `GET`; um `POST` que responde fragmento se testa com
`AssertFragment`. As opções valem só para o pedido YPY. O pedido que atualiza a página
a partir de um stream (`ypy-refresh` sem seletor) é um pedido YPY que quer o
conteúdo completo, que o `ypytest.Full()` monta: o `AssertFullPage` na resposta dele
confere que o handler não manda um fragmento pelado ali.

As demais verificações comparam os cabeçalhos `YPY-*` da resposta que conduzem a
troca, com o valor exato: `AssertRetarget`, `AssertReswap`, `AssertReselect`,
`AssertPushURL`, `AssertReplaceURL`, `AssertRedirect`, `AssertLocation` e
`AssertRefresh`.

### Testando um refresh e uma lista

```go
func TestOrderRefreshesTheCart(t *testing.T) {
	room := ypy.NewBroadcaster()
	defer room.Close()
	s := ypytest.OpenStream(t, room, ypytest.NewRequest("GET", "/room/events"))
	defer s.Close()
	s.WaitStarted()

	room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))

	got, ok := s.Next().Refresh() // Selector, Debounce (250 ms por padrão) e Jitter
	if !ok || got.Selector != "#cart" || got.Jitter != 2*time.Second {
		t.Fatalf("refresh = %+v", got)
	}
}
```

Para um handler que escreve alguns eventos e termina, basta
`res.AssertRefreshEvent(t, "#cart")` (`""` para a página) na resposta gravada. O patch
de uma lista se decodifica com `ev.Elements()`: `p.Key` e `p.Limit` trazem os valores
de `WithKey` e `WithLimit`.

## Vetores de conformidade

`spec/vectors/` no repositório guarda os vetores com que o SDK e a extensão
`ypy-stream` são testados. Cada um é um arquivo de texto com seções abertas por
`=== nome`:

| Seção | Conteúdo |
|---|---|
| `about` | uma linha, o que o vetor mostra |
| `page` | o HTML do contêiner (`<div id="room" ypy-stream="/room" ypy-history-elt>`) antes do stream |
| `input` | os bytes do stream, exatamente como um servidor os escreve |
| `dom` | o HTML do contêiner depois |
| `signals`, `events` | os patches de signals e os `ypy-event` que o cliente deve ter aplicado, em ordem |
| `respond`, `requests` | as URLs que o cliente pode pedir (`/cart => HTML`; `@page` é a própria página) e os pedidos que ele deve fazer, como `GET /cart` |

O `index.txt` lista os vetores: os patches simples (`elements-by-id`,
`elements-modes`, `elements-svg`, `signals`, `event`), as listas
(`list-append-limit`, `list-prepend-limit`, `list-key`, `list-key-limit`), o refresh
(`refresh-selector`, `refresh-page`, `refresh-debounce`, `refresh-jitter`) e
`unknown-things`, que diz que o cliente ignora um evento, uma chave ou um campo que
não conhece e que uma linha `id:` não muda nada. Só o `input` é obrigatório.

Cada lado confere uma coisa:

- **O SDK** (`vectors_test.go`) monta o stream de cada vetor com as próprias funções
  e compara os bytes com o `input`. Depois de uma mudança deliberada no formato,
  `go test -run TestVectors -update` reescreve as seções `input`; leia o diff. O teste
  também falha quando `index.txt`, os arquivos e os construtores não batem, e quando
  um evento fica sem linha `data`, que um cliente SSE não entregaria.
- **O cliente** passa o `input` ao `ypy-stream`, um byte por vez e de uma vez só, e
  compara a página, os signals, os eventos e os pedidos com as outras seções. Um
  evento ou uma chave nova no protocolo vem com um vetor.

Um servidor em outra linguagem (ou em Go sem este SDK) usa os vetores como suíte de
conformidade: para cada nome do `index.txt`, escreva o stream que o vetor descreve e
compare com o `input`. A conferência é a mesma em qualquer linguagem:

```go
func TestMyServerAgreesWithTheVectors(t *testing.T) {
	names, _ := os.ReadFile("vectors/index.txt") // uma cópia de spec/vectors
	for _, name := range strings.Fields(string(names)) {
		want := section(filepath.Join("vectors", name+".txt"), "input") // o texto depois de "=== input"
		got := myServerStream(name)                                    // os bytes que o seu servidor escreve nesse caso
		if got != want {
			t.Errorf("%s:\n got %q\nwant %q", name, got, want)
		}
	}
}
```

Use a linha `about` como descrição do que escrever; os construtores Go de
`vectors_test.go` mostram a chamada que produz cada um. Os dois lados compartilham os
arquivos por cópia (`spec/vectors` é copiado para os fixtures do cliente), então um
vetor nunca descreve algo que só um deles faz.
