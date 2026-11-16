---
title: "Documentação"
description: "Guias curtos do SDK Go do YPY. A referência da API é a documentação do pacote (go doc github.com/ypy-develop/go); o formato do streaming está em spec/PROTOCOL.md. English version: ../README.md."
order: 10
---
# Documentação

Guias curtos do SDK Go do YPY. A referência da API é a documentação do pacote
(`go doc github.com/ypy-develop/go`); o formato do streaming está em
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md). English version:
[`../README.md`](../README.md). Streaming (com listas e `ypy-refresh`), signals, cabeçalhos das extensões, proteção CSRF,
o cabeçalho `X-Robots-Tag` dos fragmentos, testes e vetores de conformidade estão
no [guia do SDK](guide.md).

## Requisições e respostas

- `ypy.FromRequest(r)` devolve um `ypy.Request`: `IsYPY`, `Boosted`,
  `HistoryRestore`, `Target`, `Source`, `RequestType` (`full` ou `partial`),
  `CurrentURL`, `Prompt` (já decodificado), `Validate` e `Layer`. `IsPartial()`
  diz se a resposta deve ser um fragmento, `IsValidation()` se é uma checagem de
  campo do `ypy-validate` e `InLayer()` se o pedido preenche uma camada.
- `ypy.Middleware` guarda isso no contexto (`ypy.FromContext`) e acrescenta
  `Vary: YPY-Request`. Use em toda rota que pode responder fragmento e página.
- Uma restauração de histórico chega **sem** `YPY-Request` e com
  `YPY-History-Restore-Request: true`: responda a página inteira.

| Função | O que faz |
|---|---|
| `Redirect(w, r, url)` | `YPY-Redirect` e 204 para requisição YPY, 303 nas demais |
| `Location(w, LocationOptions{...})` | `YPY-Location` (um caminho, ou JSON quando há opções) |
| `PushURL`, `ReplaceURL` | `YPY-Push-Url`, `YPY-Replace-Url` (`"false"` impede) |
| `Refresh` | `YPY-Refresh: true` |
| `Reswap`, `Retarget`, `Reselect` | `YPY-Reswap`, `YPY-Retarget`, `YPY-Reselect` |
| `Trigger`, `TriggerAfterSwap`, `TriggerAfterSettle` | `YPY-Trigger*`; objeto JSON nome para detalhe, só ASCII; chamadas se somam |
| `OpenLayer(w, url, LayerOptions{...})`, `OpenLayerHTML` | `YPY-Open-Layer`: só a URL, ou JSON ASCII quando há opções; modo e lado desconhecidos são recusados |
| `ExpireCache(w, padrões...)`, `KeepCache` | `YPY-Cache-Expire`: padrões separados por vírgula (chamadas se somam), ou `false` |
| `Fragment`, `FragmentStatus` | um fragmento HTML, com `X-Robots-Tag: noindex` e `Vary: YPY-Request` ([por quê](#fragmentos-e-buscadores)) |
| `Partial(w, PartialSpec{Target, Swap, ID}, markup)` | um `<ypy-partial>`, atributos escapados; mesmos cabeçalhos do `Fragment` |
| `ValidationError` | 422 com o formulário e os erros; mesmos cabeçalhos do `Fragment` |
| `SeeOther` | 303 com `Location` |
| `NoContent` | 204 |

O markup é escrito como veio: escape os dados do usuário ao montá-lo.

### Validação e camadas

Uma checagem de campo do `ypy-validate` é um `POST` com os valores do formulário
inteiro e `YPY-Validate: <campo>`, em geral para o próprio handler do formulário.
Teste `q.IsValidation()` primeiro, responda com `ValidationError` (422) ou
`NoContent` e volte antes de a ação rodar. Um pedido que preenche uma camada traz
`YPY-Layer`; acrescente `Vary: YPY-Layer` se a resposta depender dele. O
`OpenLayer` abre uma camada pelo servidor, e o `ExpireCache` vai na resposta que a
página recebe (depois de um 303 o cliente só vê a resposta final).

### Fragmentos e buscadores

Um fragmento é um pedaço de página. Se um robô de busca chegar a um (por um cache
que ignora o `Vary` ou por um link vazado), ele não deve indexá-lo. Por isso
`Fragment`, `FragmentStatus`, `ValidationError` e `Partial` acrescentam
`X-Robots-Tag: noindex` e `Vary: YPY-Request`, e o `ypy.Middleware` acrescenta o
mesmo `noindex` à resposta de uma requisição YPY que espera um fragmento
(`IsPartial()`). O cabeçalho não vai quando:

- o markup é um documento inteiro (começa com doctype ou com um elemento `<html>`);
- o handler já definiu `X-Robots-Tag` (o valor dele é mantido);
- você passa `ypy.AllowIndex()` ao helper, ou usa
  `ypy.NewMiddleware(ypy.AllowIndex())` no lugar de `ypy.Middleware`. O
  `Vary: YPY-Request` continua.

`ypy.HeaderRobotsTag` e `ypy.RobotsNoIndex` são as constantes. O cliente não lê o
cabeçalho. Teste com `res.AssertNoIndex(t)`.

## Streams

`NewStream(w, r, opts...)` inicia o fluxo e devolve erro se o writer não consegue
fazer flush. Opções: `WithKeepAlive` (15 s), `WithRetry`, `WithWriteTimeout`
(10 s, renovado a cada escrita, então um `WriteTimeout` do servidor não corta o
fluxo). Sempre `defer s.Close()`. `StreamFunc` envolve uma função como handler e
faz isso por você.

Eventos: `PatchElements(markup, opts...)` com `WithMode`, `WithSelector`,
`WithNamespace`, `WithViewTransition`; `PatchSignals(patch, WithOnlyIfMissing())`;
`Redirect(url, WithRedirectMode(...))`; `Event(nome, detalhe, WithEventTarget(sel))`;
`Refresh(opções...)` com `WithRefreshSelector`, `WithRefreshDebounce` (250 ms) e
`WithRefreshJitter`. `WithKey(attr)` e `WithLimit(n)` fazem de um patch `append` ou
`prepend` uma lista com chave e tamanho máximo. Os mesmos construtores existem como
valores (`PatchElementsEvent`, `RefreshEvent`, ...) para `Send` e `Broadcaster.Publish`. Valores de uma linha com quebra de linha são
recusados, o que impede a injeção de quadros.

Opções do `Broadcaster`: `WithSnapshot` (o estado completo para cada conexão
nova), `WithBuffer` (fila por cliente, 32), `WithStreamOptions`. Publique depois
de atualizar o seu estado. `Close` desconecta todos e espera os handlers. `Broadcaster.Refresh`,
`PatchElements`, `PatchSignals`, `Redirect` e `Event` publicam para todos os clientes.

## Signals

`ReadSignals[T]` procura no parâmetro `ypy-signals` da URL, depois no corpo JSON,
depois no campo de formulário `ypy-signals` (uma constante, `ypy.SignalsParam`).
O corpo JSON é devolvido ao request, então pode ser lido de novo.
`PatchSignals(w, patch)` responde `application/json`.

### Signals assinados

`ypy.NewSigner(chave, anteriores...)` cria um `Signer` (chaves de 32 bytes ou
mais, copiadas; assina com a primeira e confere com todas). `Bind(id)` amarra as
assinaturas a uma sessão ou usuário. A assinatura é `v1.` mais o HMAC-SHA256 em
base64url do vínculo, do nome do signal e do JSON canônico (RFC 8785) do valor,
então a ordem das chaves e o formato dos números não importam (números passam por
float64: valores acima de 2^53 vão como texto). As assinaturas ficam no signal
reservado `ypySig` (objeto de nome do signal para assinatura), que o merge patch
atualiza junto com o valor.

| | |
|---|---|
| `Sign(nome, valor)`, `Verify(nome, valor, assinatura)` | um signal |
| `SignPatch(patch, nomes...)` | o patch mais `ypySig`; envie com `PatchSignals`, `Stream.PatchSignals` ou `Broadcaster.PatchSignals`. `null` apaga a assinatura também; nomes `_locais` não são assinados |
| `PatchSignedSignals(w, signer, patch, nomes...)` | `SignPatch` e `PatchSignals` |
| `Declaration(valores, nomes...)` | o valor de um atributo `ypy-signals.server` (escape como atributo) |
| `VerifySignals(doc, nomes...)` | confere o documento que `RawSignals` devolve; sem nomes é erro, signal ou assinatura ausente é falha |
| `ReadSignedSignals[T](r, signer, nomes...)` | `ReadSignals` que confere antes |

As falhas são `*SignatureError` (campo `Signal`) com `ErrSignatureMissing` ou
`ErrSignatureInvalid`; responda 403. Rotação: publique com a chave nova primeiro e
a antiga depois, e tire a antiga quando nenhuma página guardar valores dela. A
assinatura não expira: assine uma versão ou um horário junto com o valor se o
reenvio importar. O cliente faz merge de objetos, então prefira texto, números,
booleanos e listas. A especificação é a seção 3.3 do `spec/PROTOCOL.md`.

## CSRF

`ypy.NewCSRF(chave, opções...)` devolve um `*CSRF` (imutável, seguro entre
goroutines). O token é `v1.` mais o HMAC-SHA256 em base64url, com a chave do
servidor, do id da sessão: o token de uma sessão não vale em outra e nada é guardado.

| Chamada | O que faz |
|---|---|
| `Middleware(next)` | recusa requisição não segura com 403 e corpo em texto (`ErrCSRFMissing`, `ErrCSRFInvalid`, `ErrCSRFOrigin`), dá uma sessão a toda requisição e acrescenta `Vary: Cookie` |
| `Check(r)` | a mesma checagem como erro, para proteger um handler só |
| `Token(r)`, `TokenFor(sessão)` | o token da sessão da requisição, ou de um id de sessão da aplicação |
| `Meta(r)`, `Field(r)` | `<meta name="ypy-csrf" ...>` (com `data-header` se o cabeçalho não for `X-CSRF-Token`) e um input oculto `csrf_token`, como `template.HTML` |
| `Renew(w, r)`, `RenewFor(w, sessão)` | escrevem `YPY-CSRF`, e a extensão `ypy-csrf` atualiza o meta |

Opções: `CSRFSession(fn)` (o id da sessão da aplicação; vazio cai no cookie do
middleware), `CSRFPreviousKeys(chaves...)`, `CSRFHeaderName(nome)`, `CSRFSecure()`,
`CSRFTrustedOrigins(origens...)`, `CSRFOnFailure(fn)`.

O token vem do cabeçalho (`X-CSRF-Token`) ou, se faltar, do campo `csrf_token`. O
`Origin` precisa ser o host da própria requisição ou uma origem confiável; sem
`Origin`, `Sec-Fetch-Site: cross-site` é recusado. Atrás de um proxy que reescreve
o `Host`, liste a origem pública em `CSRFTrustedOrigins`.

Um 403 por token inválido leva `YPY-CSRF` com o token da sessão atual (quando a
origem é aceita), e um token feito com chave anterior passa e recebe o atual na
resposta: o cliente se atualiza nos dois casos. Depois de um login chame
`RenewFor(w, novoIDDeSessão)`. Nos testes, `ypytest.CSRFToken(token)` manda o
cabeçalho. O formato está na seção 3.4 do [`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

## Testes

`ypytest.NewRequest(método, url, opções...)`: `Target`, `Source`, `CurrentURL`,
`Prompt`, `Boosted`, `Full`, `HistoryRestore`, `NotYPY`, `Header`, `Form`,
`Query`, `Body`, `Signals`, `SignedSignals`, `Validate`, `Layer`.

`ypytest.Do(handler, req)` devolve uma `Response` com `Status`, `Header`, `Body`,
`Document()` e as verificações `AssertStatus`, `AssertHeader`, `AssertNoHeader`,
`AssertVary`, `AssertRedirect`, `AssertSeeOther`, `AssertTrigger`,
`AssertSelector`, `AssertText`, `AssertContains`, `AssertPartial`,
`AssertSignals`, `AssertStream`, `AssertPatchElements`, `AssertRedirectEvent`,
`AssertDispatchedEvent`, `AssertRefreshEvent`, `AssertNoIndex`, `AssertOpenLayer`
(com `OpenLayer(t)` para as opções), `AssertCacheExpire`, `AssertCacheKept` e as de
fragmento, página e cabeçalhos abaixo. Os seletores CSS aceitos: tipo, `#id`, `.classe`,
atributos, os quatro combinadores, listas e as pseudoclasses `:first-child`,
`:last-child`, `:nth-child(N)`, `:empty` e `:not()`. O resto é erro, nunca uma
falha silenciosa.

Cada evento de um stream se decodifica com `Elements()` (modo, seletor, `Key`,
`Limit`, elementos), `Signals()`, `Redirect()`, `Dispatched()` e `Refresh()`
(seletor, debounce, jitter).

Para um handler que não termina, `ypytest.OpenStream(t, handler, req)` o executa
e oferece `Next()`, `NextWithin(d)`, `Comments()` e `Close()`; o teste falha se o
handler não voltar depois do `Close` (goroutine vazada).

O parser é pequeno e tolerante, feito para o markup que um handler escreve. Usa
só a biblioteca padrão, o que mantém o módulo sem dependências.

### Fragmento ou página inteira

Uma requisição YPY recebe um fragmento, sem layout. Uma requisição de navegador,
e a que restaura a página após voltar ou avançar, deve receber a página inteira,
e as duas respostas devem enviar `Vary: YPY-Request`. O defeito clássico é o
handler devolver o fragmento quando o usuário dá refresh numa URL que o cliente
empurrou no histórico: aparece uma lista pelada, sem layout.

```go
func TestItemsPage(t *testing.T) {
	fragment, page := ypytest.AssertFragmentAndPage(t, handler, "/items",
		ypytest.Target("#items"))
	fragment.AssertSelector(t, "li", 3)
	page.AssertText(t, "title", "Items")
}
```

`AssertFragmentAndPage(t, h, url, opts...)` faz três GET: um YPY (com `opts`),
um comum e um de restauração de histórico. Falha se a primeira resposta não for
fragmento, se as outras duas não forem páginas inteiras ou se alguma não listar
`YPY-Request` em `Vary`. Devolve o fragmento e a página. Para uma resposta só,
use `res.AssertFragment(t)` (sem doctype, sem elemento `html`, `head` ou `body`;
corpo vazio passa) e `res.AssertFullPage(t)` (com elementos `html` e `body`). Cada
uma existe também como função: `ypytest.AssertFragment(t, res)`.

### Cabeçalhos de troca, alvo e navegação

Comparam o cabeçalho `YPY-*` da resposta com o valor exato; cabeçalho ausente
falha com uma mensagem que diz isso. Cada uma é método de `Response` e função
(`ypytest.AssertRetarget(t, res, "#errors")`): `AssertRetarget` (`YPY-Retarget`),
`AssertReswap` (`YPY-Reswap`), `AssertReselect` (`YPY-Reselect`), `AssertPushURL`
(`YPY-Push-Url`; `"false"` confere que o push é impedido), `AssertReplaceURL`
(`YPY-Replace-Url`), `AssertRedirect` (`YPY-Redirect`, sem `Location`),
`AssertLocation` (`YPY-Location`, em texto ou JSON) e `AssertRefresh`
(`YPY-Refresh: true`). `res.Location(t)` devolve o `ypy.LocationOptions` inteiro.

## Vetores de conformidade

Os arquivos `spec/vectors/*.txt` do repositório são os bytes de um stream e a página
que eles devem produzir, compartilhados pelo SDK e pela extensão `ypy-stream`. O
`index.txt` lista todos; a seção 11 do `spec/PROTOCOL.md` define o formato, e o
[guia](guide.md#vetores-de-conformidade) mostra como um servidor em outra linguagem
se confere com eles.
