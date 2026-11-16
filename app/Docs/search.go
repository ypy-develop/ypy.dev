package docs

import (
	"sort"
	"strings"
	"unicode"
)

// Entry is one thing the search can find: a page of the manual, or a name of
// the reference tables.
type Entry struct {
	Title string `json:"t"`
	URL   string `json:"u"`
	// Where says what kind of page it is, for the reader: "Guides", "Attributes".
	Where       string   `json:"s"`
	Description string   `json:"d,omitempty"`
	Headings    []string `json:"h,omitempty"`
	// Text is the start of the body. It is cut short so the index stays small;
	// the title, the headings and the description carry what people look for.
	Text string `json:"x,omitempty"`
}

// textLimit is how much of a page's body goes into the index.
const textLimit = 500

// Index is the list of everything the search can find.
type Index []Entry

// Entries is the search index of the manual, in address order. English and
// Portuguese pages are both in it; a result says which it is.
func (m *Manual) Entries() Index {
	out := make(Index, 0, len(m.pages))
	for _, p := range m.pages {
		heads := make([]string, 0, len(p.Headings))
		for _, h := range p.Headings {
			heads = append(heads, h.Text)
		}
		text := strings.Join(strings.Fields(p.Text), " ")
		if len(text) > textLimit {
			text = text[:textLimit]
		}
		where := groupTitle(p.Group, p.Lang)
		if p.Group == "" {
			where = p.Section.Title
		}
		if p.Lang == Portuguese {
			where += " (pt-BR)"
		}
		out = append(out, Entry{Title: p.Title, URL: p.Route, Where: where, Description: p.Description, Headings: heads, Text: text})
	}
	return out
}

// Result is an entry and how well it matched.
type Result struct {
	Entry
	Score int
}

// Search ranks the entries for a query: every word has to be found somewhere,
// and a word in the title counts for more than one in a heading, which counts
// for more than one in the text. It is the same ranking the page script uses,
// so a visitor with and without JavaScript sees the same order.
func (idx Index) Search(query string, limit int) []Result {
	words := tokens(query)
	if len(words) == 0 {
		return nil
	}
	var out []Result
	for _, e := range idx {
		score := 0
		matched := true
		title := strings.ToLower(e.Title)
		heads := strings.ToLower(strings.Join(e.Headings, " "))
		desc := strings.ToLower(e.Description)
		text := strings.ToLower(e.Text)
		for _, w := range words {
			s := 0
			switch {
			case title == w:
				s = 40
			case strings.HasPrefix(title, w):
				s = 24
			case strings.Contains(title, w):
				s = 16
			}
			if strings.Contains(heads, w) {
				s += 6
			}
			if strings.Contains(desc, w) {
				s += 3
			}
			if strings.Contains(text, w) {
				s++
			}
			if s == 0 {
				matched = false
				break
			}
			score += s
		}
		if matched {
			out = append(out, Result{Entry: e, Score: score})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func tokens(q string) []string {
	return strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return unicode.IsSpace(r)
	})
}
