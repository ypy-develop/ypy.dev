package docs

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	markdown "github.com/hyz-is/arandu-markdown"

	examples "github.com/ypy-develop/ypy.dev/app/Examples"
)

// render turns the body into HTML. It runs once per page, at start.
func (p *Page) render(m *Manual, md *markdown.Module) {
	// A demo marker becomes a paragraph the renderer keeps; the controller swaps
	// it for the demo at request time (app/Examples).
	doc := md.Render(examples.Prepare(markersToText(m.rewriteLinks(p, p.body))))
	p.HTML = decorateCode(string(doc.HTML()))
	p.Headings = doc.Headings
	p.Text = examples.Unmark(doc.Text)
	p.Words = doc.Words
	p.Minutes = doc.Minutes
}

// rewriteLinks points a link between two files at the page the second became.
//
// A relative link to a .md file is resolved against the folder of the file that
// holds it, looked up among the pages, and written as the address of the page
// with its #fragment kept. A link to a file that is not a page is written as
// its words alone and listed in Broken, so a page never offers a link that
// answers 404. Links to other sites, to anchors and to mail are left as they are.
func (m *Manual) rewriteLinks(p *Page, body string) string {
	var out strings.Builder
	fenced := false
	for i, line := range strings.Split(body, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
		}
		if fenced || !strings.Contains(line, "](") {
			out.WriteString(line)
			continue
		}
		out.WriteString(m.rewriteLine(p, line))
	}
	return out.String()
}

func (m *Manual) rewriteLine(p *Page, line string) string {
	// Link text is kept when a link has to go, and it sits between "[" and the
	// match, so the replacement works on the whole bracket pair.
	return linkWithText.ReplaceAllStringFunc(line, func(match string) string {
		sub := linkWithText.FindStringSubmatch(match)
		text, target, title := sub[1], sub[2], sub[3]
		if strings.HasPrefix(match, "![") {
			return match
		}
		resolved, keep := m.resolve(p, target)
		if !keep {
			m.broken = append(m.broken, BrokenLink{From: p.File, Target: target})
			return text
		}
		return "[" + text + "](" + resolved + title + ")"
	})
}

// linkWithText matches [text](target "title") and ![alt](src). The text may hold
// one level of brackets, which is enough for `x[0]` in a link label.
var linkWithText = regexp.MustCompile(`!?\[((?:[^\[\]]|\[[^\[\]]*\])*)\]\(([^)\s]+)((?:\s+"[^"]*")?)\)`)

// resolve answers the address a link target should have, and false when the
// link has to be dropped.
func (m *Manual) resolve(from *Page, target string) (string, bool) {
	switch {
	case strings.HasPrefix(target, "#"),
		strings.HasPrefix(target, "mailto:"),
		strings.HasPrefix(target, "/"):
		return target, true
	}
	if u, err := url.Parse(target); err != nil || u.Scheme != "" || u.Host != "" {
		return target, true
	}

	file, fragment, _ := strings.Cut(target, "#")
	if !strings.HasSuffix(file, ".md") {
		// A link to a folder or to a file that is not a page: not ours to follow.
		return "", false
	}
	resolved := path.Join(path.Dir(from.File), file)
	page, ok := m.byFile[resolved]
	if !ok {
		// The file may have moved: find the one page of that name in the same
		// documentation and language. The link works, and the file that holds it
		// is listed in Moved so its author can write the new path.
		page, ok = m.moved(from, resolved)
		if !ok {
			return "", false
		}
		m.relocated = append(m.relocated, BrokenLink{From: from.File, Target: target})
	}
	if fragment != "" {
		return page.Route + "#" + fragment, true
	}
	return page.Route, true
}

var codeBlock = regexp.MustCompile(`(?s)<pre><code(?: class="language-([A-Za-z0-9_+.#-]+)")?>(.*?)</code></pre>`)

// decorateCode colours each code block and puts a copy button on it.
//
// The colours are chroma's classes, never inline styles: the site's policy is
// style-src 'self'. The text to copy is carried in an attribute, which is what
// the page script of the framework reads.
func decorateCode(body string) string {
	return codeBlock.ReplaceAllStringFunc(body, func(match string) string {
		sub := codeBlock.FindStringSubmatch(match)
		language, escaped := sub[1], sub[2]
		source := html.UnescapeString(escaped)

		label := language
		if label == "" {
			label = "text"
		}
		var b strings.Builder
		b.WriteString(`<div class="doc-code">`)
		b.WriteString(`<div class="doc-code-bar"><span class="doc-code-lang">` + html.EscapeString(label) + `</span>`)
		b.WriteString(`<button type="button" class="doc-code-copy" data-copy data-copy-text="` + html.EscapeString(source) + `">`)
		b.WriteString(`<span data-copy-label="idle">Copy</span><span data-copy-label="done">Copied</span></button></div>`)
		b.WriteString(`<pre class="chroma"><code`)
		if language != "" {
			b.WriteString(` class="language-` + language + `"`)
		}
		b.WriteString(`>` + highlight(source, language) + `</code></pre></div>`)
		return b.String()
	})
}

// highlight colours source as language. A language chroma does not know comes
// back escaped and unmarked: code with no colour is readable, and code with the
// wrong colour says something false about it.
func highlight(source, language string) string {
	source = strings.TrimSuffix(source, "\n")
	if language == "" || language == "text" {
		return html.EscapeString(source)
	}
	lexer := lexers.Get(language)
	if lexer == nil {
		return html.EscapeString(source)
	}
	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, source)
	if err != nil {
		return html.EscapeString(source)
	}
	var out bytes.Buffer
	formatter := chromahtml.New(chromahtml.WithClasses(true), chromahtml.PreventSurroundingPre(true))
	if err := formatter.Format(&out, styles.Fallback, iterator); err != nil {
		return html.EscapeString(source)
	}
	return out.String()
}

// moved finds the page a link meant when its file is no longer at that path: the
// only page with the same file name in the same section and language, or the
// index of a folder that has the name of the file.
func (m *Manual) moved(from *Page, resolved string) (*Page, bool) {
	name := path.Base(resolved)
	var found *Page
	for _, p := range m.pages {
		if p.Section.ID != from.Section.ID || p.Lang != from.Lang || path.Base(p.File) != name {
			continue
		}
		if found != nil {
			return nil, false
		}
		found = p
	}
	if found != nil {
		return found, true
	}
	// "x/index.md" asked for, and the folder is now a page of its own, or the reverse.
	stem := strings.TrimSuffix(name, ".md")
	if stem == "index" || stem == "README" {
		if p, ok := m.byFile[path.Join(path.Dir(resolved), "index.md")]; ok {
			return p, true
		}
		if p, ok := m.byFile[path.Join(path.Dir(resolved), "README.md")]; ok {
			return p, true
		}
	}
	return nil, false
}

// A marker is a line of its own, `<!-- reference:attributes -->`, where a page
// wants a table generated from YPY's name tables. The Markdown module drops
// comments, so the line becomes a paragraph holding a token, and Fill swaps the
// rendered paragraph for the table once the site has the tables.
var (
	markerLine  = regexp.MustCompile(`(?m)^[ \t]*<!--\s*reference:([a-z-]+)\s*-->[ \t]*$`)
	markerToken = regexp.MustCompile(`<p>@@reference:([a-z-]+)@@</p>`)
	markerText  = regexp.MustCompile(`@@reference:[a-z-]+@@`)
)

func markersToText(body string) string {
	return markerLine.ReplaceAllString(body, "\n@@reference:$1@@\n")
}

// Fill replaces every marker of every page with what table answers for its
// name. A name table does not know is an error: a page that asks for a table
// that does not exist stops the start instead of drawing a hole.
func (m *Manual) Fill(table func(name string) (string, error)) error {
	for _, p := range m.pages {
		var failed error
		p.HTML = markerToken.ReplaceAllStringFunc(p.HTML, func(match string) string {
			name := markerToken.FindStringSubmatch(match)[1]
			html, err := table(name)
			if err != nil && failed == nil {
				failed = fmt.Errorf("docs: %s: %w", p.File, err)
			}
			return html
		})
		if failed != nil {
			return failed
		}
		p.Text = markerText.ReplaceAllString(p.Text, "")
	}
	return nil
}
