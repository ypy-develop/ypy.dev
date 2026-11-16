---
title: "Documentação do YPY"
description: "O YPY é uma biblioteca hipermídia: um elemento HTML diz, em atributos, qual requisição fazer, quando, e onde colocar o HTML que o servidor responde."
order: 10
id: "DOC-index-pt-BR"
type: "index"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/index", "lang/pt-BR"]
---

# Documentação do YPY

O YPY é uma biblioteca hipermídia: um elemento HTML diz, em atributos, qual requisição fazer, quando, e onde colocar o
HTML que o servidor responde. Sem etapa de build para você, sem dependências, uma tag `<script>`. Inglês:
[YPY documentation](../index.md).

Esta página é o mapa. Ache a sua tarefa na tabela, siga o **caminho recomendado** (leia as páginas nessa ordem) e use
as seções abaixo quando precisar da lista completa.

**Idioma das páginas.** As páginas traduzidas ficam em `pt-BR/` e espelham os caminhos do inglês. Onde ainda não há
tradução, o link leva à página em inglês e traz a marca "(em inglês)". O código e os identificadores ficam sempre em
inglês.

## Caminho recomendado, por tarefa

| Eu quero... | Leia, nesta ordem |
|---|---|
| **experimentar o YPY em dez minutos** | [Primeiros passos](guides/getting-started.md) (HTML puro, ou Go e Arandu) |
| **entender o modelo** | [Conceitos centrais](guides/core-concepts.md), [Events](../reference/events/index.md) (em inglês) |
| **montar um formulário que valida** | [Forms](../guides/forms.md) (em inglês), [`ypy-error-*`](../reference/attributes/ypy-error.md), a [receita](recipes/go.md) da sua pilha |
| **atualizar vários lugares com uma resposta** | [Multiple targets](../guides/multiple-targets.md), [`<ypy-partial>`](../reference/tags/ypy-partial.md) (em inglês) |
| **fazer o botão voltar e o cache funcionarem** | [History and caching](../guides/history-and-caching.md), [History](../reference/history.md) (em inglês) |
| **servir o YPY pelo meu backend** | a receita: [Go](recipes/go.md), [Arandu](recipes/arandu.md), [Django](../server/recipes/django.md), [Rails](../server/recipes/rails.md) ou [Laravel](../server/recipes/laravel.md) (as três últimas em inglês) |
| **levar um projeto existente para o YPY** | [Guia de migração](guides/migration.md), depois [`ypy migrate`](../migration/migrate.md) e [`ypy check`](../migration/upgrade-check.md) (em inglês) |
| **publicar com segurança** | [Checklist de segurança](guides/security-checklist.md), [Security](../reference/security.md), [Accessibility](../guides/accessibility.md) (em inglês) |
| **deixar rápido** | [Performance](../essays/performance.md), [o plano de desempenho](../performance/plan.md) (em inglês) |
| **receber atualizações ao vivo do servidor (SSE)** | [Streaming](../extensions/stream.md), [signals](../extensions/signals.md) (em inglês) |
| **ter autocompletar e tipos no meu editor** | [Editor support](../reference/editors.md) (em inglês) |
| **usar ou escrever uma extensão** | [Extensions](../extensions/using-and-writing.md), [Official extensions](../extensions/index.md) (em inglês) |
| **usar uma biblioteca de componentes (Bootstrap, React, Alpine, Web Component)** | [Third-party components](../guides/third-party-components.md), [`ypy-preserve`](../reference/attributes/ypy-preserve.md) (em inglês) |
| **testar o meu servidor sem navegador** | [Testing your app](../server/testing.md) (em inglês) |
| **trabalhar no próprio YPY** | [Core map](../contributing/core-map.md), ADRs (em inglês) |

## Começar

- [Primeiros passos](guides/getting-started.md): a primeira página com HTML puro, o primeiro servidor com Go, o Arandu.
- [Conceitos centrais](guides/core-concepts.md): atributos, gatilhos, alvos, trocas e morph, parciais, cabeçalhos,
  herança, estilização dos estados com o Tailwind.

## Conceitos

| Página | O que explica |
|---|---|
| [Conceitos centrais](guides/core-concepts.md) | o modelo inteiro numa página: atributos, gatilhos, trocas, cabeçalhos |
| [Events](../reference/events/index.md) (em inglês) | a vida de uma requisição, todos os eventos, `detail.ctx` |
| [Swap styles and morph](../reference/swap-and-morph.md) (em inglês) | como uma resposta entra, e o que as mudanças do próprio usuário preservam |
| [Attribute inheritance](../reference/inheritance.md) (em inglês) | `:inherited` explícito, `:append` |
| [History](../reference/history.md) (em inglês) | a barra de endereço, voltar e avançar |

## Guias

| Guia | Para |
|---|---|
| [Primeiros passos](guides/getting-started.md) | os primeiros dez minutos |
| [Conceitos centrais](guides/core-concepts.md) | o vocabulário |
| [Checklist de segurança](guides/security-checklist.md) | antes de publicar |
| [Guia de migração](guides/migration.md) | de marcação legada para o YPY |
| [Events](../reference/events/index.md), [Forms](../guides/forms.md), [Multiple targets](../guides/multiple-targets.md), [History and caching](../guides/history-and-caching.md), [Accessibility](../guides/accessibility.md), [Third-party components](../guides/third-party-components.md), [Extensions](../extensions/using-and-writing.md), [Performance](../essays/performance.md), [Testing your app](../server/testing.md) | os demais guias, por enquanto só em inglês |
| [`ypy migrate`](../migration/migrate.md), [`ypy check`](../migration/upgrade-check.md) | as duas ferramentas de migração (em inglês) |

## Referência

A referência está em inglês: [atributos](../reference/attributes/ypy-error.md) ·
[`<ypy-partial>`](../reference/tags/ypy-partial.md) · [eventos de fila](../reference/events/request-queue.md) ·
[eventos de alvo](../reference/events/target.md) · [offline e online](../reference/events/ypy-offline.md) ·
[`config.classes`](../reference/config/classes.md) · [segurança](../reference/security.md) ·
[signals](../extensions/signals.md) · [extensões oficiais](../extensions/index.md) ·
[swap e morph](../reference/swap-and-morph.md) · [histórico](../reference/history.md) ·
[herança](../reference/inheritance.md) · [streaming `ypy-stream`](../extensions/stream.md) · [suporte a editores](../reference/editors.md). A lista completa, por assunto, está no
[índice em inglês](../index.md#reference).

## Receitas

Cada receita é um servidor funcionando, com os cabeçalhos `YPY-*`, `422` e `303`, `Vary: YPY-Request`, CSRF, marcação
com Tailwind e um teste.

| Pilha | Receita |
|---|---|
| Go, `net/http` | [Go](recipes/go.md) |
| Arandu (Hesape, Kyse) | [Arandu](recipes/arandu.md) |
| Django | [Django](../server/recipes/django.md) (em inglês) |
| Rails | [Rails](../server/recipes/rails.md) (em inglês) |
| Laravel | [Laravel](../server/recipes/laravel.md) (em inglês) |

## Migração

[Guia de migração](guides/migration.md): o caminho, o passo a passo, o que muda no comportamento e as tabelas de equivalência de atributos,
eventos, cabeçalhos, configuração, extensões e API JavaScript. As ferramentas: o [`ypy migrate`](../migration/migrate.md)
converte, o [`ypy check`](../migration/upgrade-check.md) lista o que se comporta de outro jeito no YPY. Há um só comportamento e nenhuma camada de
compatibilidade (ADR-0019).

## Segurança

[Checklist de segurança](guides/security-checklist.md) (faça antes de publicar) ·
[Security](../reference/security.md) (os ajustes, `callbacks`, o portão por nonce, Trusted Types, a CSP recomendada) ·
[Auditoria de 2026-11-15](../security/audit-2026-11-15.md) · [Accessibility](../guides/accessibility.md).

## Desempenho

[Performance](../essays/performance.md) · [Plano e medições](../performance/plan.md) ·
[Linha de base](../performance/baseline.md) · `ypy bench`.

## Ferramentas Go: `ypy`

Um binário, subcomandos explícitos, sem Node. Rode `go run ./cmd/ypy <comando>` no repositório, ou instale.

| Comando | O que faz | Leia |
|---|---|---|
| `ypy build` | gera `dist/`: pacotes, arquivos minificados, extensões, cópias Brotli; `--sizes` impõe os orçamentos (núcleo 18 kB, extensão 6 kB) | [Primeiros passos](guides/getting-started.md#pegue-o-arquivo) |
| `ypy test` | roda as suítes de navegador em Chrome, Firefox e Safari; `--dist` roda contra os arquivos compilados | [Testing your app](../server/testing.md#what-ypytest-cannot-see) (em inglês) |
| `ypy ci` | roda todos os portões: nomes, licenças, SPDX, formatação, vet, testes, tamanhos, guia, navegadores, migração | [Testing your app](../server/testing.md#what-ypytest-cannot-see) (em inglês) |
| `ypy bench` | mede o desempenho num navegador de verdade, compara dois resultados, faz perfil | [Performance](../essays/performance.md#measure) (em inglês) |
| `ypy check` | acha usos antigos que quebram ou mudam | [`ypy check`](../migration/upgrade-check.md) (em inglês) |
| `ypy gen` | gera os dados de editor (VS Code, JetBrains, JSON Schema) a partir da tabela de nomes; `--check` falha se um arquivo estiver desatualizado | [Editor support](../reference/editors.md) (em inglês) |
| `ypy migrate` | converte um projeto escrito antes do YPY | [`ypy migrate`](../migration/migrate.md) (em inglês) |

## SDK Go

O lado do servidor do YPY para Go: `github.com/ypy-develop/go`. Ele lê a requisição (`ypy.FromRequest`,
`ypy.Middleware`), responde fragmentos, elementos `<ypy-partial>`, redirecionamentos e os cabeçalhos de resposta
`YPY-*`, lê e aplica patches de signals, transmite mudanças por eventos enviados pelo servidor a um cliente ou a muitos,
e testa tudo isso sem navegador (`ypytest`). Só biblioteca padrão. Comece por [Primeiros
passos](guides/getting-started.md#caminho-b-go-e-arandu), pela [receita Go](recipes/go.md) e por
[Testing your app](../server/testing.md) (em inglês); a documentação do próprio SDK é o README, as
[páginas do SDK Go](../../go/README.md) e o `spec/PROTOCOL.md` dele.

## Convenções destas páginas

- Todo exemplo usa classes do Tailwind CSS; nada no YPY depende dele.
- Os nomes são `ypy-*` e `data-ypy-*` (atributos), `YPY-*` (cabeçalhos), `ypy:*` (eventos), `ypy` (o objeto global).
- Cada página abre com um front matter curto (`id`, `type`, `status`, `updated`, `tags`), como as notas do workspace, e
  liga às outras páginas por caminhos relativos.
- A documentação é escrita em inglês e em português do Brasil; as páginas em português espelham os caminhos em
  [`pt-BR/`](index.md).
