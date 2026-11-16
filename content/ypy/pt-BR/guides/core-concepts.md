---
title: "Conceitos centrais"
description: "O modelo inteiro cabe numa página: um elemento leva atributos, um gatilho dispara, o YPY faz uma requisição, o servidor responde com HTML e uma troca (swap) coloca o HTML na página."
order: 20
id: "GUIDE-core-concepts-pt-BR"
type: "guide"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/guide", "lang/pt-BR"]
---

# Conceitos centrais

O modelo inteiro cabe numa página: um elemento leva **atributos**, um **gatilho** dispara, o YPY faz uma
**requisição**, o servidor responde com HTML e uma **troca** (swap) coloca o HTML na página. **Eventos** marcam cada
passo. Inglês: [Core concepts](../../guides/core-concepts.md).

Conteúdo: [Atributos](#atributos) · [Gatilhos](#gatilhos) · [Alvos](#alvos) · [Trocas e morph](#trocas-e-morph) ·
[Parciais e trocas fora de ordem](#parciais-e-trocas-fora-de-ordem) · [Requisições e cabeçalhos](#requisições-e-cabeçalhos) ·
[Herança](#herança) · [Padrões](#padrões) · [Eventos](#eventos) · [Estilizar os estados](#estilizar-os-estados) ·
[Melhoria progressiva](#melhoria-progressiva)

## Atributos

Todo atributo se escreve `ypy-<nome>`. A grafia `data-ypy-<nome>` significa exatamente o mesmo (use-a se o seu
validador de HTML reclamar). Se um elemento tiver as duas, vale o `ypy-*`.

| Grupo | Atributos | O que dizem |
|---|---|---|
| Requisição | `ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` | o verbo e a URL |
| Quando | `ypy-trigger` | qual evento e seus modificadores |
| Onde | `ypy-target`, `ypy-select`, `ypy-swap`, `ypy-swap-oob`, `ypy-select-oob` | para onde vai a resposta e como |
| O que é enviado | `ypy-vals`, `ypy-include`, `ypy-headers`, `ypy-encoding`, `ypy-validate` | valores extras, valores de outros elementos, cabeçalhos |
| Histórico | `ypy-push-url`, `ypy-replace-url`, `ypy-boost` | a barra de endereço e a navegação de página inteira |
| Retorno visual | `ypy-indicator`, `ypy-disable`, `ypy-confirm`, `ypy-busy`, `ypy-announce` | indicadores, bloqueio, pergunta, leitores de tela |
| Concorrência | `ypy-sync` | o que fazer quando chega uma segunda requisição |
| Erros | `ypy-error-target`, `ypy-error-swap`, `ypy-error-focus`, `ypy-fallback`, `ypy-retry` | para onde vão as falhas |
| Morph | `ypy-preserve`, `ypy-preserve-attrs`, `ypy-morph-force`, `ypy-focus`, `ypy-scroll`, `ypy-main` | o que a fusão preserva |
| Código | `ypy-on:<evento>` | um handler (desligado por padrão, veja [Segurança](security-checklist.md)) |
| Configuração | `<meta name="ypy-config" content='...'>`, `ypy-config` | ajustes da página ou de uma requisição |

Existe uma página de referência (em inglês) para os menos óbvios: [`ypy-error-*`](../../reference/attributes/ypy-error.md),
[`ypy-sync`](../../reference/attributes/ypy-sync.md), [`ypy-preserve`](../../reference/attributes/ypy-preserve.md),
[`ypy-focus`](../../reference/attributes/ypy-focus.md), [`ypy-scroll`](../../reference/attributes/ypy-scroll.md),
[`ypy-main`](../../reference/attributes/ypy-main.md), [`ypy-retry`](../../extensions/retry.md),
[`ypy-announce`](../../reference/attributes/ypy-announce.md), [`ypy-busy`](../../reference/attributes/ypy-busy.md).

```html
<form ypy-post="/subscribe" ypy-target="#box" ypy-swap="outerMorph" ypy-indicator="#spin"
      class="flex items-center gap-2">
  <input name="email" type="email" required class="rounded border p-2">
  <button class="rounded bg-sky-600 px-3 py-2 text-white">Assinar</button>
  <span id="spin" class="ypy-indicator text-sm text-slate-500">Enviando...</span>
</form>
```

## Gatilhos

`ypy-trigger` nomeia o evento que inicia a requisição. Sem ele, o YPY escolhe o natural: `submit` para um formulário,
`change` para input, select ou textarea, `click` para o resto.

```html
<!-- ao carregar, e depois a cada 30 segundos -->
<div ypy-get="/stats" ypy-trigger="load, every 30s" class="rounded border p-4"></div>

<!-- busca ativa: espera 300 ms depois da última tecla, só quando o valor mudou -->
<input name="q" ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms, search" class="rounded border p-2">

<!-- carrega a próxima página quando esta linha entra na tela -->
<tr ypy-get="/rows?page=2" ypy-trigger="revealed" ypy-swap="afterend"></tr>
```

| Gatilho | Dispara |
|---|---|
| qualquer evento do DOM (`click`, `input`, `change`, `submit`, `keyup`, ...) | quando o evento acontece no elemento |
| `load` | uma vez, quando o elemento é processado |
| `revealed` | uma vez, na primeira vez que entra na tela |
| `intersect` | toda vez que entra na área visível (`root:`, `threshold:`) |
| `every 5s` | num temporizador, enquanto o elemento estiver na página |

| Modificador | Efeito |
|---|---|
| `once` | só na primeira vez |
| `changed` | só se o valor mudou desde a última requisição |
| `delay:300ms` | espera; um novo evento reinicia a espera (debounce) |
| `throttle:1s` | no máximo uma vez por intervalo |
| `from:<seletor>` | escuta em outro elemento (`from:document`, `from:closest form`) |
| `target:<seletor>` | só para eventos que começaram num elemento que casa |
| `consume` | impede que o evento suba para os ancestrais |
| `[expressão]` | uma condição; compila código, então exige `callbacks` ligado (veja [Segurança](security-checklist.md#código-do-autor)) |

## Alvos

`ypy-target` diz para onde vai a resposta. Aceita um seletor CSS ou uma das formas relativas:

| Valor | Elemento |
|---|---|
| `#list`, `.panel`, `main` | o primeiro que casa no documento |
| `this` | o elemento que enviou a requisição |
| `closest form` | o ancestral mais próximo que casa |
| `find .row` | o primeiro descendente que casa |
| `next .panel`, `previous .panel` | o irmão seguinte ou anterior que casa |

Sem `ypy-target`, vale o próprio elemento. Quando o seletor não casa com nada, o YPY avisa no console e dispara
[`ypy:target:missing`](../../reference/events/target.md) em vez de falhar em silêncio. `ypy-select` escolhe *parte da
resposta* por seletor (`ypy-select="#results"` pega só aquele elemento de uma página inteira).

## Trocas e morph

`ypy-swap` diz como a resposta entra. **O padrão é um morph**: a resposta é fundida na página, e os
elementos que existem dos dois lados continuam sendo os mesmos nós do DOM, então foco, texto digitado, posição de
rolagem e `<details>` abertos sobrevivem.

| Estilo | O que faz |
|---|---|
| `innerMorph` (padrão) | funde a resposta nos filhos do alvo |
| `outerMorph` | funde a resposta no próprio alvo |
| `innerHTML`, `outerHTML` | substitui os filhos, ou o alvo (cru, nada é preservado) |
| `beforeend`, `afterbegin`, `beforebegin`, `afterend` | insere ao lado ou dentro do alvo |
| `textContent`, `delete`, `none` | define texto, remove o alvo, não faz nada |

Os modificadores vêm depois do estilo: `ypy-swap="innerHTML swap:200ms settle:100ms show:top"`. As regras completas, a
ordem das operações e o que significa "o usuário vence" estão em
[Swap styles and morph](../../reference/swap-and-morph.md) (em inglês).

## Parciais e trocas fora de ordem

Uma resposta pode atualizar vários lugares. Ponha cada pedaço extra num `<ypy-partial>`; o cliente troca primeiro o
conteúdo principal e depois cada parcial no seu próprio alvo:

```html
<li class="rounded border p-2">Novo item</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">3 itens</ypy-partial>
<ypy-partial id="toast"><p class="rounded bg-emerald-600 p-3 text-white">Salvo</p></ypy-partial>
```

Sem `ypy-target`, a parcial substitui o elemento da página que tem o mesmo `id`. Veja
[`<ypy-partial>`](../../reference/tags/ypy-partial.md) e [Multiple targets](../../guides/multiple-targets.md) para os
padrões (ambos em inglês).

## Requisições e cabeçalhos

O YPY envia cabeçalhos para o servidor distinguir uma requisição de fragmento de um carregamento de página, e o
servidor responde com cabeçalhos que dirigem o cliente. Os nomes são `YPY-*`.

| Cabeçalho da requisição | Valor |
|---|---|
| `YPY-Request` | `true` em toda requisição do cliente |
| `YPY-Target` | o alvo como `tag#id`, por exemplo `ul#list` |
| `YPY-Source` | o elemento que fez a requisição, como `tag#id` |
| `YPY-Request-Type` | `partial` (um alvo menor) ou `full` (o conteúdo da página é substituído) |
| `YPY-Current-URL` | a URL da página |
| `YPY-Boosted` | `true` num link ou formulário "boosted" |
| `YPY-History-Restore-Request` | `true` ao restaurar a página no voltar ou avançar (`YPY-Request` fica de fora, para responderem a página inteira) |
| `YPY-Prompt` | a resposta de um prompt (extensão) |

| Cabeçalho da resposta | Efeito |
|---|---|
| `YPY-Redirect` | navega com carga de página inteira |
| `YPY-Location` | navega com uma requisição do YPY, sem carga inteira |
| `YPY-Push-Url`, `YPY-Replace-Url` | mudam a barra de endereço (`false` impede) |
| `YPY-Refresh` | recarrega a página |
| `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect` | mudam onde, como e qual parte é trocada |
| `YPY-Trigger` | dispara eventos no cliente quando a resposta chega (um nome, uma lista ou um objeto JSON com detalhes). Para agir depois da troca, escute `ypy:after:swap` |

Duas regras que todo servidor segue:

- **Responda um fragmento só a uma requisição do YPY, e uma página inteira a todo o resto**, e acrescente
  `Vary: YPY-Request` para um cache manter os dois separados.
- **Use os códigos de status que já significam algo**: `422` para formulário inválido, `303` depois de um `POST`
  que deu certo, `204` para "nada a trocar". [Forms](../../guides/forms.md) (em inglês) tem os detalhes.

## Herança

Um atributo só chega aos elementos de baixo **quando você manda**, com `:inherited`:

```html
<section ypy-target:inherited="#results" ypy-swap:inherited="innerHTML">
  <button ypy-get="/people">Pessoas</button>
  <button ypy-get="/places">Lugares</button>
</section>
<div id="results"></div>
```

Nada é herdado por acidente, que é a surpresa mais comum dos projetos escritos antes do YPY. Regras, `:append` e a tabela dos
atributos: [Attribute inheritance](../../reference/inheritance.md) (em inglês).

## Padrões

O YPY tem um só comportamento, e seus padrões são os seguros: morph por padrão, herança explícita, um `422` vai para
dentro do formulário, código do autor (`ypy-on`, valores `js:`) fica desligado, `aria-busy` é marcado enquanto uma
requisição roda. Um projeto escrito antes do YPY que dependa de outro comportamento é convertido com `ypy migrate`, e
a [Migração](migration.md) lista o que muda.

## Eventos

Cada passo da vida de uma requisição dispara um evento do DOM chamado `ypy:<substantivo>:<verbo>` que sobe pela
árvore: `ypy:before:request`, `ypy:after:swap`, `ypy:error`... Escute com `addEventListener`, com
`ypy-on::before:request` num elemento, ou numa extensão. A lista, a ordem e o `detail` estão em
[Events](../../reference/events/index.md) (em inglês).

## Estilizar os estados

O YPY nunca injeta estilo em linha e não impõe CSS. Ele informa o estado de duas formas, para você estilizar com
qualquer sistema de CSS:

| Estado | Classe (configurável em [`config.classes`](../../reference/config/classes.md)) | Atributo |
|---|---|---|
| requisição em andamento | `ypy-request` no elemento e nos seus indicadores | nenhum |
| trocando, assentando, conteúdo novo | `ypy-swapping`, `ypy-settling`, `ypy-added` no alvo | nenhum |
| a última requisição falhou | | `data-ypy-state="error"` no formulário ou elemento, até a próxima requisição |
| sem conexão | `ypy-offline` em `<html>` | `data-ypy-state="offline"` em `<html>` |

Com o **Tailwind CSS v4**, transforme essas classes em variantes uma vez, na sua folha de estilo:

```css
@import "tailwindcss";

/* escaneie também os templates do servidor, ou as classes de um fragmento somem do build */
@source "../templates";
@source "../**/*.go";

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
@custom-variant ypy-offline (.ypy-offline &);
```

e use-as na marcação:

```html
<button ypy-post="/save"
        class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50 ypy-request:cursor-wait">
  Salvar
</button>

<form ypy-post="/signup" class="rounded border p-4 ypy-error:border-red-500">...</form>

<p class="hidden ypy-offline:block text-amber-700">Você está sem conexão.</p>
```

O padrão do indicador, um ícone de carregamento que só aparece durante a requisição:

```html
<span id="spin" class="ypy-indicator hidden ypy-request:inline">Carregando...</span>
```

Para Bootstrap, Bulma e os outros, estilize as mesmas classes e leia
[Third-party components](../../guides/third-party-components.md) (em inglês) para reiniciar os scripts deles depois de
uma troca.

## Melhoria progressiva

Construa a página para que links e formulários funcionem sem script: um `<form method="post" action="/items">` e um
`<a href="/items/3">`. Depois acrescente os atributos `ypy-*`. O servidor usa `YPY-Request` para responder um
fragmento ao YPY, e uma página, ou um redirecionamento `303`, ao navegador comum. Se o script não carregar, o site
continua funcionando.

## Para onde ir

[Forms](../../guides/forms.md) · [History and caching](../../guides/history-and-caching.md) ·
[Accessibility](../../guides/accessibility.md) · [Checklist de segurança](security-checklist.md) ·
[Índice da documentação](../index.md)
