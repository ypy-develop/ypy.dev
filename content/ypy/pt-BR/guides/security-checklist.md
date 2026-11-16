---
title: "Checklist de segurança"
description: "Percorra esta lista antes de publicar. Cada item é algo que você faz ou verifica; as razões e os ajustes estão em Security (em inglês), e os achados da última revisão estão na auditoria."
order: 20
id: "GUIDE-security-checklist-pt-BR"
type: "guide"
status: "draft"
lang: "pt-BR"
updated: "2026-11-15"
tags: ["type/guide", "lang/pt-BR", "topic/security"]
---

# Checklist de segurança

Percorra esta lista antes de publicar. Cada item é algo que você faz ou verifica; as razões e os ajustes estão em
[Security](../../reference/security.md) (em inglês), e os achados da última revisão estão na
[auditoria](../../security/audit-2026-11-15.md). Inglês: [Security checklist](../../guides/security-checklist.md).

A ideia por trás de tudo: **o YPY transforma atributos em comportamento**, então um atacante que consiga escrever
marcação nas suas páginas (conteúdo guardado, um parâmetro refletido) não precisa de script; um `ypy-post` num botão
basta. Três camadas respondem a isso, e nenhuma substitui o escape no servidor.

## 1. O servidor escapa tudo

- [ ] Todo valor que você põe no HTML passa pelo escape do motor de templates (`html/template`, Jinja, ERB, Blade,
      Twig, templ). Fragmentos são HTML como qualquer outro.
- [ ] Texto do usuário nunca cai **dentro de uma tag** que o YPY lê (`<div ` + dado do usuário) nem num valor de
      atributo que começa com `ypy-` (`ypy-get="` + dado do usuário). Escape os valores de atributo; não deixe o
      usuário escolher o nome de um atributo.
- [ ] Conteúdo do usuário (comentários, perfis) é renderizado numa região da página sobre a qual você consegue
      raciocinar, higienizado no servidor, e sem poder carregar atributos `ypy-*` nem `data-ypy-*`. Remova-os no seu
      higienizador.
- [ ] As respostas ao YPY são `Content-Type: text/html; charset=utf-8` e `X-Content-Type-Options: nosniff`.
- [ ] Os valores que você manda em `YPY-Trigger`, `YPY-Location` ou `YPY-Redirect` vêm do seu código, não da
      requisição. Um destino de redirecionamento tirado de um parâmetro de URL é um redirecionamento aberto: aceite só
      caminhos do seu próprio site.

## 2. O código do autor fica desligado

- [ ] A página mantém o padrão `callbacks:"none"` e não define `callbacks:"all"` (veja
      [Código do autor](#código-do-autor) abaixo). Handlers de um projeto convertido por `ypy migrate` não rodam até
      você decidir: mover a lógica para um script, ou escolher `"nonce"`.
- [ ] Nenhum handler `ypy-on:...`, nenhum `js:` em `ypy-vals`/`ypy-confirm`, nenhum `[condição]` em `ypy-trigger`, a
      não ser que você tenha escolhido um dos dois caminhos seguros:
  - `callbacks:"nonce"` com a extensão `ypy-csp` e `safeEval:true`, ou
  - mover a lógica para um arquivo de script que escuta eventos ([Events](../../reference/events/index.md), em inglês).

## 3. Uma Content Security Policy estrita

- [ ] Envie uma política sem `unsafe-eval` e sem `unsafe-inline`:

```
Content-Security-Policy:
  default-src 'self';
  script-src 'nonce-{{ nonce }}' 'strict-dynamic';
  object-src 'none';
  base-uri 'none';
  connect-src 'self';
  require-trusted-types-for 'script';
  trusted-types ypy
```

- [ ] `connect-src 'self'` está presente: é o segundo cadeado atrás da regra de mesma origem do próprio YPY.
- [ ] Acrescente a `trusted-types` os nomes de outras políticas de Trusted Types que a sua página criar.
- [ ] Nenhum atributo `style` em linha escrito por você também; o YPY não escreve nenhum. O Tailwind e uma folha de
      estilo não precisam.

## 4. Requisições de mesma origem

- [ ] `config.mode` fica em `"same-origin"` (o padrão). Mude só numa página que realmente
      fala com outra origem, e configure CORS nesse servidor.
- [ ] Nenhum endpoint aceita mudança de estado por `GET`. `ypy-get` é para ler; `ypy-post`, `ypy-put`,
      `ypy-patch`, `ypy-delete` mudam coisas.

## 5. CSRF

Um site autenticado por cookie precisa de proteção contra requisições entre sites. Escolha a que o seu framework dá e
confira que **as requisições do YPY levam o token**:

- [ ] Formulários levam o token num campo oculto (o helper de formulário do framework faz isso). Isso cobre
      `<form ypy-post>`.
- [ ] Requisições que nenhum formulário sustenta (`ypy-delete` num botão, `ypy-patch` num interruptor) levam o token
      num cabeçalho. Ponha-o uma vez no layout, ou acrescente-o a toda requisição com um único ouvinte:

```html
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ csrfToken }}"}'>
```

```js
document.addEventListener('ypy:config:request', (event) => {
  event.detail.ctx.request.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]').content
})
```

- [ ] Os cookies são `SameSite=Lax` (ou `Strict`), `Secure` e `HttpOnly` na sessão.
- [ ] Um token expirado responde um status que a página consegue mostrar (419 ou 403), e o layout recarrega ou explica.
      Cada [receita](../recipes/go.md) mostra o nome do cabeçalho que o seu framework lê.

## 6. Caches e `Vary`

- [ ] Toda rota que responde um fragmento ao YPY e uma página a um navegador envia `Vary: YPY-Request`. Um cache que
      ignore isso pode servir o fragmento como página inteira, ou o fragmento de um usuário a outro.
- [ ] Fragmentos privados enviam `Cache-Control: private, no-store` (ou `private` com vida curta), nunca um cabeçalho
      de cache público.
- [ ] Uma requisição de restauração de histórico leva `YPY-History-Restore-Request: true` e **não** leva
      `YPY-Request`; o servidor responde uma página inteira ([History](../../reference/history.md), em inglês).

## 7. Trave a configuração

- [ ] Carregue `ypy-lock` e defina `locked:true`, listando `ypy-lock` em `extensions`, depois de escrever os ajustes
      que você quer. A tag `<meta name="ypy-config">` só vale dentro de `<head>`, então mantenha dados do usuário fora
      do head.
- [ ] `extensions` lista só os nomes que você entrega ([Extensions](../../extensions/using-and-writing.md#restrict), em inglês).

```html
<meta name="ypy-config" content='extensions:"ypy-lock,ypy-csp",nonce:true,callbacks:"nonce",locked:true'>
```

A trava impede código descuidado e injeção de marcação no head. Ela não impede um script que já roda na sua página:
para isso existe a CSP.

## 8. Portão por nonce, para páginas que mostram marcação do usuário

- [ ] Quando usuários podem pôr marcação numa página, ligue o `ypy-csp` com `nonce:true`. O servidor carimba cada
      elemento do YPY que renderiza com o nonce da página (`ypy-nonce="{{ nonce }}"`); elementos sem ele perdem os
      atributos `ypy-*`.
- [ ] Um nonce novo em cada resposta, nunca uma constante, nunca numa página em cache.
- [ ] `inlineScriptNonce` fica sem valor, a não ser que você confie nos scripts das suas próprias respostas.

## 9. Streams e sockets

- [ ] A rota por trás de `ypy-sse:connect` ou `ypy-ws:connect` confere a sessão como qualquer outra rota, e confere o
      `Origin` do upgrade do WebSocket.
- [ ] Um broadcaster envia a cada usuário só o que ele pode ver; filtre por conexão, não no cliente.
- [ ] Não ponha segredos na URL do stream nem em `ypy-vals`: os dois aparecem em logs e no código-fonte da página.

## 10. Verifique

- [ ] `ypy check` e os testes de segurança dos seus próprios handlers passam
      ([Testing your app](../../server/testing.md), em inglês): escreva um que posta um payload com marcação `ypy-post`
      e confere que ele volta escapado.
- [ ] Abra a página com o console ligado: `ypy: author code is off` e eventos `ypy:security:blocked` dizem que algo
      tentou executar código; cada um é um achado a explicar.
- [ ] Rode os testes contra os arquivos compilados: `go run ./cmd/ypy test --dist`.

## Código do autor

Quatro lugares compilam texto de um atributo em código: handlers `ypy-on`, valores `js:` (`ypy-vals`, `ypy-headers`,
`ypy-confirm`), a `[condição]` de `ypy-trigger` e as expressões de `ypy-live`. Uma função decide, com
`config.callbacks`:

| Valor | Significado |
|---|---|
| `"none"` | nada é compilado. O padrão |
| `"nonce"` | o código só roda pelo `ypy-csp`, depois de ele conferir o nonce do elemento |
| `"all"` | o código roda. Nunca é o padrão |

Qualquer outro valor conta como `"none"`: um erro de digitação não consegue ligar código. Quando o código é bloqueado,
um aviso é registrado e `ypy:security:blocked` dispara no elemento com `{reason, code}`; o filtro dá falso, o confirm
nega, o valor `js:` é descartado.

## Veja também

[Security reference](../../reference/security.md) · [Auditoria 2026-11-15](../../security/audit-2026-11-15.md) ·
[Extensions](../../extensions/using-and-writing.md) · [Migração](migration.md) · [Receitas](../recipes/go.md)
