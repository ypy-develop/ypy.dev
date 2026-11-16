package site

import (
	"encoding/json"

	"github.com/arandu-io/hesape/geo"
)

// What the structured data says about who made YPY: facts the footer and the
// header already show, and the two profiles the project owns.
const (
	publisherName = "HYZIS"
	publisherURL  = "https://hyz.is"
	// Repository is where the code of YPY is.
	Repository = "https://github.com/ypy-develop/ypy"
	// LinkedIn is the project's page there.
	LinkedIn   = "https://www.linkedin.com/showcase/ypy-dev"
	licenseURL = "https://opensource.org/licenses/MIT"
)

// Crumb is one step of the trail above a page.
type Crumb struct {
	Name string
	Href string
}

func ref(id string) map[string]any { return map[string]any{"@id": id} }

// graph turns the module's two-entity graph (WebSite and WebPage) into the
// page's one graph: the site with its search, the publisher, the page itself
// and, by kind of page, the software, the article and the trail. Every fact is
// one the page shows; there is no rating, no offer and no review. A page that
// asked not to be indexed has no graph, and gets none.
func (s *Site) graph(meta geo.Metadata, in PageInfo) geo.Metadata {
	if len(meta.Schema) == 0 {
		return meta
	}
	var doc struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal(meta.Schema, &doc); err != nil {
		return meta
	}

	origin := s.URL
	websiteID := origin + "#website"
	orgID := origin + "#organization"
	pageID := meta.Canonical + "#webpage"
	updated := parseDate(in.Updated)

	var image map[string]any
	if meta.Image != "" {
		image = map[string]any{
			"@type": "ImageObject", "url": meta.Image,
			"width": meta.ImageWidth, "height": meta.ImageHeight, "caption": meta.ImageAlt,
		}
	}

	for _, node := range doc.Graph {
		switch node["@type"] {
		case "WebSite":
			// The home page's canonical address ends in a slash, so the site's does.
			node["url"] = origin + "/"
			node["inLanguage"] = "en"
			node["publisher"] = ref(orgID)
			node["potentialAction"] = map[string]any{
				"@type":       "SearchAction",
				"target":      map[string]any{"@type": "EntryPoint", "urlTemplate": origin + "/search?q={search_term_string}"},
				"query-input": "required name=search_term_string",
			}
		case "WebPage":
			node["isPartOf"] = ref(websiteID)
			if image != nil {
				node["primaryImageOfPage"] = image
			}
			if !updated.IsZero() {
				node["dateModified"] = updated.Format("2006-01-02")
			}
			if len(in.Crumbs) > 0 {
				node["breadcrumb"] = ref(meta.Canonical + "#breadcrumb")
			}
			if in.Kind == KindHome {
				node["about"] = ref(origin + "#software")
			}
		}
	}

	doc.Graph = append(doc.Graph, map[string]any{
		"@type":  "Organization",
		"@id":    orgID,
		"name":   publisherName,
		"url":    publisherURL,
		"sameAs": []string{GitHub, LinkedIn},
	})

	switch in.Kind {
	case KindHome:
		version := s.YPYVersion()
		software := map[string]any{
			"@type":               []string{"SoftwareSourceCode", "SoftwareApplication"},
			"@id":                 origin + "#software",
			"name":                s.Name,
			"description":         HomeDescription,
			"url":                 origin + "/",
			"applicationCategory": "DeveloperApplication",
			"license":             licenseURL,
			"codeRepository":      Repository,
			"programmingLanguage": "JavaScript",
			"version":             version,
			"softwareVersion":     version,
			"author":              ref(orgID),
			"publisher":           ref(orgID),
		}
		if meta.Image != "" {
			software["image"] = meta.Image
		}
		doc.Graph = append(doc.Graph, software)
	case KindArticle:
		article := map[string]any{
			"@type":            "TechArticle",
			"@id":              meta.Canonical + "#article",
			"headline":         in.Title,
			"description":      in.Description,
			"url":              meta.Canonical,
			"inLanguage":       meta.Language,
			"mainEntityOfPage": ref(pageID),
			"isPartOf":         ref(websiteID),
			"author":           ref(orgID),
			"publisher":        ref(orgID),
		}
		if !updated.IsZero() {
			article["dateModified"] = updated.Format("2006-01-02")
		}
		if meta.Image != "" {
			article["image"] = meta.Image
		}
		doc.Graph = append(doc.Graph, article)
	}

	if len(in.Crumbs) > 0 {
		items := make([]map[string]any, len(in.Crumbs))
		for i, c := range in.Crumbs {
			items[i] = map[string]any{"@type": "ListItem", "position": i + 1, "name": c.Name, "item": origin + c.Href}
		}
		doc.Graph = append(doc.Graph, map[string]any{
			"@type": "BreadcrumbList", "@id": meta.Canonical + "#breadcrumb", "itemListElement": items,
		})
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return meta
	}
	meta.Schema = body
	return meta
}
