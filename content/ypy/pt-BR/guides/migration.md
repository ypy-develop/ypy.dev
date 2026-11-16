---
title: "Guia de migração: de marcação legada para o YPY"
description: "Levar um projeto escrito antes do YPY leva minutos para converter e uma revisão para terminar."
order: 20
id: "GUIDE-migration-pt-BR"
type: "guide"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/guide", "lang/pt-BR", "topic/migration"]
---

# Guia de migração: de marcação legada para o YPY

Levar um projeto escrito antes do YPY leva minutos para converter e uma revisão para terminar. Um comando converte os
nomes, um arquivo de script substitui o antigo, e uma lista de trabalho mostra onde o YPY se comporta diferente. O YPY
tem um só comportamento e nenhuma camada de compatibilidade: um projeto convertido roda com `ypy.min.js` desde o
início. Inglês: [Migration guide](../../migration/index.md).

Este é o **caminho e as tabelas de equivalência**. As duas ferramentas têm páginas próprias, geradas ou escritas a
partir do código (em inglês): [`ypy migrate`](../../migration/migrate.md) (todas as opções e a saída JSON) e
[`ypy check`](../../migration/upgrade-check.md) (a lista de trabalho e a tabela de regras, uma linha para cada
comportamento que mudou). A decisão por trás disso é a ADR-0019.

## Como os nomes são escritos neste guia

A marcação legada escreve atributos, cabeçalhos, eventos e classes com nomes que o YPY não usa. Este guia os escreve
como marcadores, e o `ypy migrate` **aprende os nomes com o seu projeto** (ou você os passa por opção):

| Marcador | É | Formato de exemplo (inventado) |
|---|---|---|
| prefixo `<p>-` | o prefixo legado de atributo: `<p>-get`, `data-<p>-target`, `<p>-on::after-swap` | `acme-get` |
| prefixo `<P>-` | o prefixo legado de cabeçalho: `<P>-Request`, `<P>-Target` | `ACME-Request` |
| `<id>` | o identificador legado: o objeto global, o prefixo de evento, as classes de estado | `acme.process(el)`, `acme:afterSwap`, `acme-request` |

Se o seu projeto não os revela (um repositório de back end que só lê um cabeçalho, por exemplo), diga à ferramenta:
`ypy migrate --prefix acme --header-prefix ACME --id acme .`

## O caminho

| Etapa | Você faz | Tempo | Resultado |
|---|---|---|---|
| **1. Converter** | `ypy migrate --write`, depois carrega `ypy.min.js` no lugar do script antigo | minutos | os nomes são os do YPY e o projeto carrega o núcleo do YPY |
| **2. Revisar o comportamento** | resolve o que o `ypy check` lista: herança explícita, respostas de erro, código do autor, eventos | no seu ritmo, página por página | o projeto se comporta como o YPY, sem nada para desligar depois |

Não há ponte: nenhum arquivo de compatibilidade, nenhum perfil e nenhuma configuração que traga o comportamento antigo
de volta. O que muda no comportamento está [abaixo](#o-que-muda-no-comportamento) e, para os seus arquivos, no
`ypy check`.

O backend não precisa mudar o código por causa dos nomes: o YPY lê e envia os mesmos cabeçalhos com o prefixo `YPY-`, e
o seu servidor convertido pelo `ypy migrate` lê `YPY-Request` onde lia `<P>-Request`. Os cabeçalhos novos
(`YPY-Source`, `YPY-Request-Type`, ...) são adicionais.

## Passo a passo

1. **Comece com a árvore limpa.** Faça commit do que tem: a conversão é um diff comum que você pode revisar e desfazer.
2. **Leia a lista de trabalho.** `ypy check .` lista o que quebra (`quebra`) ou se comporta diferente (`muda`) além da
   grafia, com a linha, a correção e um link para a regra. Não muda nada. Cada achado traz uma gravidade:

   | Gravidade | Significado | Antes de publicar |
   |---|---|---|
   | `quebra` | para de funcionar depois da mudança para o YPY | corrija |
   | `muda` | funciona, mas se comporta diferente do que o projeto foi escrito para fazer | revise |
   | `aviso` | funciona; só a grafia é legada | o `ypy migrate` converte |

3. **Ensaio.** `ypy migrate .` mostra os prefixos aprendidos e o diff, e não escreve nada. Confira se os prefixos e o
   identificador estão certos.
4. **Escreva.** `ypy migrate --write .`, depois `ypy migrate .` de novo: deve informar `0 of N file(s) would change`.
5. **Troque o script.** Ponha `dist/ypy.min.js` onde estava o script antigo, apague o arquivo antigo e os scripts das
   extensões, e carregue de `dist/ext/` as extensões oficiais que você usava (veja a [tabela de extensões](#extensões)
   abaixo), ou use o `dist/ypy.max.min.js`, que as traz.
6. **Corrija o comportamento.** Percorra a lista do `ypy check` (veja [O que muda no
   comportamento](#o-que-muda-no-comportamento)).
7. **Rode os seus portões**: formatação, build, vet, testes unitários, testes de navegador. Um teste que afirma a
   grafia legada de um atributo ou cabeçalho foi convertido junto com o resto.
8. **Confira a olho** uma página de cada tipo (uma lista, um formulário, uma busca, um stream) no navegador com o
   console aberto.

Para uma família de projetos que dependem uns dos outros (um framework, um kit de interface e produtos), converta na
ordem das dependências: primeiro as bibliotecas que desenham os atributos, depois os produtos; a nota mais longa está
em [`ypy migrate`](../../migration/migrate.md#step-by-step-for-an-arandu-or-kyse-project).

## O que muda no comportamento

A conversão renomeia; ela não faz uma página se comportar como antes. Estas são as diferenças, e cada uma é uma regra
do `ypy check`:

| Comportamento | YPY |
|---|---|
| Troca padrão | morph (`innerMorph`): o estado vivo, como foco, seleção e rolagem, vence o HTML que chega. Use `ypy-swap="innerHTML"` onde quiser a substituição pura ([Swap styles and morph](../../reference/swap-and-morph.md), em inglês) |
| Herança de atributos | explícita: um atributo num contêiner só chega aos descendentes com `:inherited` (`ypy-target:inherited="#out"`). O `ypy migrate` não sabe de quais atributos de contêiner os descendentes dependem, então o `ypy check` os sinaliza ([Attribute inheritance](../../reference/inheritance.md), em inglês) |
| `422` de um formulário | trocado dentro do formulário ([Forms](../../guides/forms.md), em inglês) |
| Outros 4xx e 5xx | decididos pelo núcleo: alvos de erro e `ypy-error-*` ([`ypy-error`](../../reference/attributes/ypy-error.md), em inglês); `ypy-status:<código>` muda um status |
| Código do autor (`ypy-on`, valores `js:`, atributos de evento) | `callbacks: "none"` por padrão; ligue com `callbacks:"all"` sabendo o que isso significa ([Checklist de segurança](security-checklist.md#código-do-autor)) |
| `aria-busy` durante uma requisição | automático ([`ypy-busy`](../../reference/attributes/ypy-busy.md), em inglês) |
| Tempo limite padrão da requisição | 60 s (`defaultTimeout`); `0` não é "sem limite" |
| Requisições entre origens | só a mesma origem (`mode`) |
| Nomes de evento e `detail` | `ypy:<substantivo>:<verbo>`, com `detail` igual a `{ctx}` ([Events](../../reference/events/index.md#detailctx), em inglês) |
| Atributos que não existem mais | `<p>-vars`, `<p>-params`, `<p>-request`, `<p>-ext`, `<p>-inherit`, `<p>-disinherit`, `<p>-history`, o modificador `queue:` do gatilho e a forma de um handler por linha do `<p>-on` (substituições abaixo) |
| API JavaScript e configuração | as funções e os nomes de configuração legados acabaram; as substituições estão nas tabelas abaixo |

## Tabelas de equivalência

As tabelas mostram o que o `ypy migrate` faz (grafia) e o que o `ypy check` informa (comportamento). Toda linha que não
é uma simples troca de nome é uma regra da [tabela de regras](../../migration/upgrade-check.md#reference-every-rule).

### Atributos

Todo atributo mantém o nome e perde o prefixo: `<p>-get` vira `ypy-get`, `data-<p>-target` vira `data-ypy-target`,
`<p>-on::after-swap` vira `ypy-on::after-swap`. Estes são os que mudaram de **significado** ou de **disponibilidade**:

| Antes | Depois |
|---|---|
| `<p>-get`, `-post`, `-put`, `-patch`, `-delete`, `-target`, `-swap`, `-trigger`, `-select`, `-include`, `-indicator`, `-confirm`, `-push-url`, `-boost`, `-headers`, `-vals`, `-sync`, `-on:<evento>` | o mesmo nome com `ypy-` |
| `<<p>-partial>` (tag) | `<ypy-partial>` |
| `<p>-vars` | `ypy-vals` com o prefixo `js:`: `ypy-vals="js:{a: 1}"` |
| `<p>-params` | removido: filtre num ouvinte de `ypy:config:request` (`detail.ctx.request.body` é um `FormData`) |
| `<p>-request` | `ypy-config` (o mesmo JSON: `timeout`, `credentials`, `noHeaders`) |
| `<p>-disabled-elt` | `ypy-disable` |
| `<p>-disable` (sem valor: não processar) | `ypy-ignore` (faça isto **antes** de renomear o de cima) |
| `<p>-ext` | apagado: uma extensão fica ativa quando o script dela é carregado; restrinja com `config.extensions` |
| `<p>-inherit`, `<p>-disinherit` | apagados; escreva `:inherited` nos atributos que devem chegar aos descendentes |
| herança implícita de `<p>-target`, `-swap`, ... | explícita: `ypy-target:inherited="#out"` |
| `<p>-history="false"` | apagado (os instantâneos não ficam mais no `localStorage`) |
| `<p>-prompt` | a extensão `ypy-prompt` (o mesmo atributo, o mesmo cabeçalho `YPY-Prompt`) |
| modificador `queue:` do gatilho | `ypy-sync="this:queue last"` (ou `queue first`, `queue all`, `drop`, `abort`) |
| `show:<seletor>:<posição>`, `scroll:<seletor>:<posição>` | `show:top showTarget:#sel`, `scroll:bottom scrollTarget:#sel` |
| a grafia com traço de `ypy-on` (`ypy-on-click`) | a grafia com dois pontos `ypy-on:click`, `ypy-on::after:swap` |
| `<p>-on="evento: código"` (um handler por linha) | um atributo `ypy-on:<evento>` por handler, ou `ypy-on="click -> código"` |
| a palavra `inherit` em `-include` e `-indicator` | removida; use `:inherited` |

### Eventos

O prefixo `<id>:` vira `ypy:`, e os nomes seguem um só padrão, `ypy:<substantivo>:<verbo>`. O `detail` é `{ctx}`
([Events](../../reference/events/index.md#detailctx)); o legado (`xhr`, `requestConfig`, `parameters`, ...) não existe. O
`ypy migrate` converte o prefixo, e o `ypy check` lista todo evento cujo nome ou `detail` mudou, com o nome a usar.

| Antes | YPY |
|---|---|
| `<id>:beforeRequest`, `beforeSend` | `ypy:before:request` |
| `<id>:configRequest` | `ypy:config:request` |
| `<id>:afterRequest` | `ypy:after:request` (`ypy:finally:request` também na falha) |
| `<id>:beforeSwap` | `ypy:before:swap` (a decisão de trocar passou para `config.noSwap` e `ypy-status`) |
| `<id>:afterSwap`, `oobAfterSwap` | `ypy:after:swap` |
| `<id>:afterSettle` | `ypy:after:settle` |
| `<id>:load` | `ypy:after:process` |
| `<id>:beforeProcessNode`, `afterProcessNode` | `ypy:before:init`, `ypy:after:init` |
| `<id>:beforeOnLoad`, `afterOnLoad` | `ypy:before:response`, `ypy:after:request` |
| `<id>:beforeCleanupElement` | `ypy:before:cleanup` |
| `<id>:responseError` | `ypy:response:error` |
| `<id>:sendError`, `sendAbort`, `timeout`, `swapError`, `targetError` | `ypy:error` (`ypy:abort` para abortar, `ctx.errorKind` para o tipo) |
| `<id>:historyRestore`, `historyCacheMiss` | `ypy:before:history:restore` |
| `<id>:pushedIntoHistory`, `replacedInHistory` | `ypy:after:history:push`, `ypy:after:history:replace` |
| `<id>:beforeHistoryUpdate`, `beforeHistorySave` | `ypy:before:history:update` |
| `<id>:beforeTransition` | `ypy:before:viewTransition` |
| `<id>:sseOpen`, `sseMessage`, `sseError`, `sseClose` | `ypy:sse:after:connection`, `ypy:sse:after:message`, `ypy:sse:error`, `ypy:sse:close` |
| `<id>:wsOpen`, `wsClose`, `wsError` | `ypy:ws:after:connection`, `ypy:ws:close`, `ypy:ws:error` |
| `<id>:validation:*`, `<id>:xhr:*` | **acabaram**: a validação usa a API do navegador, as requisições usam `fetch`; use o evento nativo `invalid` e `ypy:before:request`, `ypy:finally:request`, `ypy:abort` (`quebra`) |
| `ypy-on::after-request` (gêmeo em kebab-case) | `ypy-on::after:request` |

A lista completa (todo evento, uma regra cada) é `ypy check --list-rules`; a tabela gerada está na
[tabela de regras](../../migration/upgrade-check.md#reference-every-rule).

### Cabeçalhos

O `ypy migrate` converte o nome onde quer que apareça no código do servidor: `<P>-Request` vira `YPY-Request`,
`Vary: <P>-Request` vira `Vary: YPY-Request`, `HTTP_<P>_REQUEST` vira `HTTP_YPY_REQUEST` (a forma CGI que servidores
PHP, Python WSGI e Ruby expõem).

| Antes | Depois | Notas |
|---|---|---|
| `<P>-Request` | `YPY-Request` | `true` em toda requisição do YPY |
| `<P>-Target` | `YPY-Target` | o valor é `tag#id` (`div#feed`), não só o id |
| `<P>-Trigger` (requisição) | `YPY-Source` | `tag#id` do elemento que fez a requisição |
| `<P>-Trigger-Name` (requisição) | `YPY-Source`, ou um campo no corpo | |
| `<P>-Current-URL`, `<P>-Boosted`, `<P>-History-Restore-Request`, `<P>-Prompt` | `YPY-Current-URL`, `YPY-Boosted`, `YPY-History-Restore-Request`, `YPY-Prompt` | o prompt precisa da extensão `ypy-prompt` |
| `<P>-Redirect`, `-Location`, `-Push-Url`, `-Replace-Url`, `-Refresh`, `-Reswap`, `-Retarget`, `-Reselect` (resposta) | o mesmo com `YPY-` | |
| `<P>-Trigger` (resposta) | `YPY-Trigger` | |
| `<P>-Trigger-After-Swap`, `-After-Settle` (resposta) | `YPY-Trigger`, e escute `ypy:after:swap` ou `ypy:after:settle` | |
| (novo) | `YPY-Request-Type`: `partial` ou `full` | uma requisição `full` significa que o conteúdo da página é substituído |
| `Vary: <P>-Request` | `Vary: YPY-Request` | mantenha em toda rota que responde uma página e um fragmento |

Uma **restauração de histórico** chega sem `YPY-Request` e com `YPY-History-Restore-Request: true`. Responda uma página
inteira.

### Configuração

Definida em `<meta name="<id>-config">` (agora `ypy-config`) ou em `<id>.config` (agora `ypy.config`). Uma chave legada
que não está nesta tabela acabou; o `ypy check` lista as que conhece.

| Antes | Depois |
|---|---|
| `defaultSwapStyle` | `defaultSwap` |
| `globalViewTransitions` | `transitions` |
| `historyEnabled` | `history` |
| `includeIndicatorStyles` | `includeIndicatorCSS` |
| `timeout` (`0` era "nenhum") | `defaultTimeout` (padrão 60000 ms) |
| `selfRequestsOnly` | `mode`: `true` é `"same-origin"`, `false` é `"cors"` |
| `disableInheritance` | removida: a herança é sempre explícita |
| `requestClass`, `indicatorClass`, `addedClass`, `settlingClass`, `swappingClass` | `classes.request`, `classes.indicator`, `classes.added`, `classes.settling`, `classes.swapping` |
| `responseHandling` | `config.noSwap` (uma lista de status) ou `ypy-status:<código>="swap:none"` |
| `withCredentials` | `ypy-config="credentials:include"` |
| `allowEval` | removida: a CSP da página decide, e `config.callbacks` |
| `allowScriptTags`, `allowNestedOobSwaps`, `attributesToSettle`, `triggerSpecsCache`, `refreshOnHistoryMiss`, `historyCacheSize`, `getCacheBusterParam`, `methodsThatUseUrlParams`, `scrollBehavior`, `scrollIntoViewOnBoost`, `defaultSwapDelay`, `disableSelector`, `ignoreTitle`, `reportValidityOfForms`, `wsBinaryType`, `wsReconnectDelay` | removidas, cada uma com a substituição na sua regra (`ypy check --list-rules`) |

### Extensões

Os arquivos e atributos das extensões são renomeados com o mesmo prefixo, e o registro delas foi substituído (veja a
tabela da API abaixo). Carregue o arquivo do YPY depois do núcleo.

| Extensão legada | YPY | Atributos |
|---|---|---|
| eventos enviados pelo servidor (`<p>-sse.js`) | `ext/ypy-sse.min.js` | `sse-connect`, `sse-close` viram `ypy-sse:connect`, `ypy-sse:close`. `sse-swap` vira `ypy-sse-swap`, que a extensão apenas avisa. São `quebra`: o modelo de troca pelo nome do evento acabou; mensagens sem nome são trocadas automaticamente e eventos com nome viram eventos do DOM |
| WebSocket | `ext/ypy-ws.min.js` | `ws-connect`, `ws-send` viram `ypy-ws:connect`, `ypy-ws:send` |
| suporte ao head | `ext/ypy-head.min.js` | (nenhum) |
| preload | `ext/ypy-preload.min.js` | `preload="mouseover"` vira `ypy-preload="mouseover"` |
| response targets | o núcleo: `ypy-status:<código>` | `ypy-status:404="target:#x"`, `ypy-status:4xx`, `ypy-status:5xx` |
| a troca com morph | o núcleo: `innerMorph`, `outerMorph` | nenhuma extensão necessária |
| um script que registra uma extensão (`defineExtension`) | `ypy.registerExtension(nome, hooks)`, baseada em eventos | reescreva: [Extensions](../../extensions/using-and-writing.md#write-one) (em inglês) |

O `ypy.max.min.js` reúne SSE, WebSocket, preload e as outras extensões comuns num arquivo só.

### A API JavaScript

| Antes | Depois |
|---|---|
| `<id>.process(el)`, `<id>.ajax(...)`, `<id>.on(...)`, `window.<id>` | `ypy.process(el)`, `ypy.ajax(...)`, `ypy.on(...)`, `window.ypy` (o `ypy migrate` converte) |
| `<id>.addClass`, `removeClass`, `toggleClass`, `takeClass`, `closest`, `remove`, `values`, `off`, `location`, `logger`, `logAll`, `logNone` | removidos: `classList`, `closest()`, `remove()`, `new FormData(form)`, `removeEventListener()`, `ypy.ajax()`, `console`, `ypy.config.logAll` (a regra nomeia a substituição) |
| `<id>.config.<chave>` | `ypy.config.<chave>`; chaves renomeadas acima |

### Classes

`<id>-request`, `<id>-indicator`, `<id>-swapping`, `<id>-settling`, `<id>-added` viram `ypy-request`, `ypy-indicator`,
`ypy-swapping`, `ypy-settling`, `ypy-added` na marcação, nos seletores e no CSS. Mude os nomes em
`ypy.config.classes` se o seu CSS usa outros ([config](../../reference/config/classes.md)). No Tailwind, veja
[Estilizar os estados](core-concepts.md#estilizar-os-estados).

### Tag de script e meta tag

| Antes | Depois |
|---|---|
| `<script src=".../<id>.min.js">` | `<script src=".../ypy.min.js">` (um arquivo local; uma URL externa é listada, não editada: use `--script-src`) |
| `<meta name="<id>-config" content='...'>` | `<meta name="ypy-config" content='...'>` |

## O que o `ypy migrate` deixa para você

- **Comentários e texto corrido** (`--comments` converte também os nomes reconhecidos nos comentários). O relatório
  final lista as linhas que ainda trazem um nome legado em palavras.
- **Tags de script externas**, uma cópia de um script legado ou de uma extensão dele no seu repositório (listadas e
  ignoradas: troque pelos arquivos do YPY), `node_modules`, `vendor`, `dist`, `build`, `.git`, arquivos minificados e
  binários e arquivos acima de 2 MB.
- **Nomes montados em tempo de execução**: `"<p>-" + nome`, um cabeçalho lido com um nome montado por partes. A
  ferramenta reconhece texto literal, então procure estes à mão.
- **Código de terceiros** que renderiza os atributos legados (um kit de interface, uma biblioteca de componentes no
  servidor): migre-o ou atualize-o primeiro, ou a saída dele não será entendida.

## Notas do servidor, por pilha

O trabalho do servidor depois da conversão é pequeno: ler `YPY-Request`, enviar `Vary: YPY-Request`, usar `422` e
`303`, e levar o token de CSRF. Cada receita tem o código funcionando:

| Pilha | Receita | O cabeçalho legado no código do servidor se parece com |
|---|---|---|
| Go (`net/http`) | [Go](../recipes/go.md) | `r.Header.Get("<P>-Request")` |
| Arandu (Hesape, Kyse) | [Arandu](../recipes/arandu.md) | os helpers do framework, convertidos junto com a biblioteca |
| Django, Flask, Jinja | [Django](../../server/recipes/django.md) (em inglês) | `request.headers.get("<P>-Request")`, `request.META["HTTP_<P>_REQUEST"]` |
| Rails | [Rails](../../server/recipes/rails.md) (em inglês) | `request.headers["<P>-Request"]` |
| Laravel, PHP | [Laravel](../../server/recipes/laravel.md) (em inglês) | `$request->header('<P>-Request')`, `$_SERVER['HTTP_<P>_REQUEST']` |

Um servidor que precisa atender clientes legados e do YPY durante a transição pode ler os dois cabeçalhos por um tempo;
o nome legado é o único lugar em que ele aparece escrito, no seu código, e some quando a última página legada sumir.

## Verifique

```sh
ypy check .                      # a lista de trabalho: nenhum "quebra" sobrando
ypy migrate .                    # "0 of N file(s) would change": a conversão está completa e estável
ypy ci names                     # no repositório do YPY: nenhum nome legado em lugar nenhum (no seu projeto, procure com grep os marcadores que a ferramenta aprendeu)
```

e a sua própria suíte de testes, num navegador, contra o arquivo que você serve agora.

## Desfazer

A conversão é um commit comum. Faça `git revert`, ponha a tag do script anterior de volta, e você está onde estava. O
YPY não muda os seus dados nem o seu servidor.

## Perguntas frequentes

**Posso manter o comportamento antigo por um tempo?** Não. Há um só comportamento, o deste guia, então a revisão do
`ypy check` faz parte da mudança. Faça página por página: converta, resolva a lista daquela página, publique.

**Meus handlers `<p>-on` legados pararam de funcionar.** O código do autor fica desligado por padrão
(`callbacks: "none"`). Para ligar: `<meta name="ypy-config" content='callbacks:"all"'>`, sabendo o que isso significa
([Checklist de segurança](security-checklist.md#código-do-autor)).

**A grafia legada e a nova podem conviver?** Num elemento, `ypy-*` vence `data-ypy-*`. Não existe grafia legada no
YPY: a conversão não é opcional.

## Veja também

[`ypy migrate`](../../migration/migrate.md) · [`ypy check`](../../migration/upgrade-check.md) ·
ADR-0019 · [Events](../../reference/events/index.md) ·
[Checklist de segurança](security-checklist.md) · [Receitas](../recipes/go.md)
