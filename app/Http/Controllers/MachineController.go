package controllers

import (
	"net/http"

	hhttp "github.com/arandu-io/hesape/http"

	site "github.com/ypy-develop/ypy.dev/app/Site"
)

// MachineController answers /llms.txt. The other three documents crawlers and
// language models ask for by name, robots.txt, sitemap.xml and llms-full.txt, are
// served by hesape/geo from the same list of pages (site.Catalog), wired in
// bootstrap/app.go; this one is the site's own because it groups the pages by
// section, which the module's flat list does not.
type MachineController struct {
	Controller
}

// NewMachineController builds it.
func NewMachineController(s *site.Site) *MachineController {
	return &MachineController{Controller{site: s}}
}

// Llms is /llms.txt. A deployment that is not indexed answers with the
// module's own "discovery disabled" text and lists no page.
func (c *MachineController) Llms(ctx *hhttp.Context) error {
	body := c.site.LlmsText()
	if !c.site.Indexing {
		body = "# Discovery disabled\n\n> This deployment does not publish a model-facing index.\n"
	}
	h := ctx.Response.Header()
	h.Set("Content-Type", "text/plain; charset=utf-8")
	h.Set("Content-Language", "en")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Robots-Tag", "noindex")
	ctx.Response.WriteHeader(http.StatusOK)
	_, err := ctx.Response.Write([]byte(body))
	return err
}
