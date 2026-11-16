// Command brand draws the images the site shares and shows in a tab, from the
// two logo files that were drawn, which it never modifies:
//
//   - assets/ypy.svg, the full logo, used on the social covers;
//   - assets/ypy-icon.svg, the icon, used as it is for the favicons.
//
// It writes, into public/:
//
//   - social-cover.png and one cover per section, 1200 by 630, from
//     resources/brand/social-cover.html;
//   - with -icons, the PNG icons (16, 32, 48, 180, 192 and 512 pixels) and
//     favicon.ico, each the icon file drawn as it is.
//
// The PNGs are rasterised by a headless Chrome or Chromium, found through the
// CHROME_BIN variable, the usual macOS location, or the PATH. The browser runs
// without a window, with a profile in a temporary folder that is removed at the
// end, and is stopped as soon as the screenshot is written. Nothing else is
// needed.
//
// It is run by hand, from the repository root, when the brand or a cover's
// words change, and its output is committed, so building the site never needs
// Chrome:
//
//	go run ./tools/brand          # the covers
//	go run ./tools/brand -icons   # the covers and the favicons
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// cover is one social image: the file it is written to, the name under the
// logo and the one line under the name.
type cover struct {
	File        string
	Name        string
	Description string
}

// covers are the words of every cover. They repeat the site's own sentences
// (the front page cards and the header), so a link preview says what the page
// says.
var covers = []cover{
	{"social-cover.png", "YPY", "HTML attributes ask the server for HTML. One script, no build step."},
	{"social-cover-docs.png", "YPY Docs", "Guides, core concepts and the server side, from a first page."},
	{"social-cover-reference.png", "YPY Reference", "Every attribute, event, header and configuration key."},
	{"social-cover-examples.png", "YPY Examples", "Click to edit, active search, bulk update, delete a row."},
	{"social-cover-extensions.png", "YPY Extensions", "Streaming, WebSocket, history cache and more, one file each."},
	{"social-cover-go.png", "YPY Go SDK", "The server side in Go: fragments, forms, streaming and tests."},
}

type raster struct {
	file string
	size int
}

// rasters are the icon sizes a browser, a phone's home screen and the web
// manifest ask for. favicon-48x48.png only feeds favicon.ico.
var rasters = []raster{
	{"favicon-16x16.png", 16},
	{"favicon-32x32.png", 32},
	{"favicon-48x48.png", 48},
	{"apple-touch-icon.png", 180},
	{"android-chrome-192x192.png", 192},
	{"android-chrome-512x512.png", 512},
}

func main() {
	log.SetFlags(0)
	icons := flag.Bool("icons", false, "also write the favicons and the PNG icons of public/")
	flag.Parse()
	if err := run(*icons); err != nil {
		log.Fatal(err)
	}
}

func run(withIcons bool) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, f := range []string{"assets/ypy.svg", "assets/ypy-icon.svg", "resources/brand/social-cover.html"} {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			return fmt.Errorf("run this from the repository root: %s was not found", f)
		}
	}
	chrome, err := findChrome()
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "ypy-brand-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	tpl, err := template.ParseFiles(filepath.Join(root, "resources", "brand", "social-cover.html"))
	if err != nil {
		return err
	}
	for _, c := range covers {
		var page bytes.Buffer
		err := tpl.Execute(&page, map[string]string{
			"Root": root, "Name": c.Name, "Description": c.Description, "Domain": "ypy.dev",
		})
		if err != nil {
			return err
		}
		img, err := render(chrome, tmp, page.String(), 1200, 630)
		if err != nil {
			return err
		}
		if err := writePNG(filepath.Join(root, "public", c.File), img); err != nil {
			return err
		}
	}

	if withIcons {
		images := map[string]image.Image{}
		for _, r := range rasters {
			page := fmt.Sprintf(`<!doctype html><meta charset="utf-8"><style>html,body{margin:0;background:transparent}img{display:block}</style>`+
				`<img src="file://%s/assets/ypy-icon.svg" width="%d" height="%d" alt="">`, root, r.size, r.size)
			img, err := render(chrome, tmp, page, r.size, r.size)
			if err != nil {
				return err
			}
			images[r.file] = img
			if r.file == "favicon-48x48.png" {
				continue
			}
			if err := writePNG(filepath.Join(root, "public", r.file), img); err != nil {
				return err
			}
		}
		ico := []image.Image{images["favicon-16x16.png"], images["favicon-32x32.png"], images["favicon-48x48.png"]}
		if err := writeICO(filepath.Join(root, "public", "favicon.ico"), ico); err != nil {
			return err
		}
	}
	fmt.Println("brand assets written")
	return nil
}

// findChrome returns a Chrome or Chromium binary.
func findChrome() (string, error) {
	if p := os.Getenv("CHROME_BIN"); p != "" {
		return p, nil
	}
	if p := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"; fileExists(p) {
		return p, nil
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no Chrome or Chromium found: set CHROME_BIN")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// render draws html in a headless browser and returns the top-left width by
// height pixels of the screenshot. The window is at least 800 pixels wide,
// because a headless window narrower than that is widened by the browser, and
// the crop is what makes the result the size that was asked for.
func render(chrome, tmp, html string, width, height int) (image.Image, error) {
	page := filepath.Join(tmp, "page.html")
	shot := filepath.Join(tmp, "shot.png")
	if err := os.WriteFile(page, []byte(html), 0o644); err != nil {
		return nil, err
	}
	_ = os.Remove(shot)
	w, h := max(width, 800), max(height, 800)
	// The profile is inside the temporary folder, so the run never touches the
	// profile of a browser somebody has open, and a mock keychain keeps a fresh
	// profile from waiting on the system's. There is no debugging port.
	cmd := exec.Command(chrome, "--headless=new", "--disable-gpu", "--hide-scrollbars", "--use-mock-keychain",
		"--user-data-dir="+filepath.Join(tmp, "profile"), "--no-first-run", "--no-default-browser-check",
		"--allow-file-access-from-files", "--force-device-scale-factor=1", "--virtual-time-budget=4000",
		fmt.Sprintf("--window-size=%d,%d", w, h), "--screenshot="+shot, "file://"+page)
	var stderr lockedBuffer
	cmd.Stderr = &stderr
	// Its own process group, so stopping it stops its helper processes too.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("chrome did not start: %w", err)
	}
	stop := func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	// Chrome reports the screenshot on stderr once the file is complete, and on
	// some systems it then keeps running. So wait for that line, with a
	// deadline, and stop the browser here either way.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	deadline := time.After(90 * time.Second)
	for written := false; !written; {
		select {
		case err := <-done:
			stop()
			if err != nil {
				return nil, fmt.Errorf("chrome failed: %w\n%s", err, stderr.String())
			}
			written = true
		case <-deadline:
			stop()
			<-done
			return nil, fmt.Errorf("chrome wrote no screenshot in 90 seconds\n%s", stderr.String())
		case <-time.After(200 * time.Millisecond):
			if strings.Contains(stderr.String(), "bytes written to file") {
				stop()
				<-done
				written = true
			}
		}
	}
	f, err := os.Open(shot)
	if err != nil {
		return nil, fmt.Errorf("chrome wrote no screenshot: %w\n%s", err, stderr.String())
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	// Copy the top-left corner into a fresh image, whose origin is (0,0) for
	// the encoders.
	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out.Set(x, y, img.At(x, y))
		}
	}
	return out, nil
}

// lockedBuffer is a buffer the browser writes from one goroutine while the
// tool reads it from another.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func writePNG(path string, img image.Image) error {
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// writeICO writes an .ico that holds PNG images, which every browser that still
// asks for /favicon.ico reads.
func writeICO(path string, imgs []image.Image) error {
	var data [][]byte
	for _, img := range imgs {
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			return err
		}
		data = append(data, b.Bytes())
	}
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, [3]uint16{0, 1, uint16(len(imgs))})
	offset := 6 + 16*len(imgs)
	for i, img := range imgs {
		bounds := img.Bounds()
		entry := struct {
			W, H, Colors, Reserved uint8
			Planes, Bits           uint16
			Size, Offset           uint32
		}{uint8(bounds.Dx()), uint8(bounds.Dy()), 0, 0, 1, 32, uint32(len(data[i])), uint32(offset)}
		_ = binary.Write(&out, binary.LittleEndian, entry)
		offset += len(data[i])
	}
	for _, d := range data {
		out.Write(d)
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}
