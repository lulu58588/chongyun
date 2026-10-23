//go:build ignore

// Icon generator. Everything the app shows is drawn from one mark: a cloud
// with a slot cut out of its base, on a 44-unit grid.
// starling.svg is the full drawing, the slot cut out of the cloud; the app
// tile uses it from 64px up.
// starling-small.svg drops the slot, which would only be a smudge; the tray
// and the smaller tiles use it, and so does the header in
// internal/gui/assets/index.html.
//
//	go run build/icon/gen.go tray internal/gui/tray.png     # 44px black template icon (macOS menu bar)
//	go run build/icon/gen.go tray-flap internal/gui/trayflap # its frames when clicked
//	go run build/icon/gen.go app 64 internal/gui/icon.png   # app icon at a given size
//	make icons                                              # regenerates all of them plus the icns
package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	//go:embed starling.svg
	fullSVG string
	//go:embed starling-small.svg
	smallSVG string
)

func main() {
	var img image.Image
	switch {
	case len(os.Args) == 3 && os.Args[1] == "tray":
		img = tray()
	case len(os.Args) == 3 && os.Args[1] == "tray-flap":
		if err := trayFlap(os.Args[2], 44); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case len(os.Args) == 4 && os.Args[1] == "tray-flap":
		n, _ := strconv.Atoi(os.Args[3])
		if err := trayFlap(os.Args[2], n); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case len(os.Args) == 4 && os.Args[1] == "glyph":
		n, _ := strconv.Atoi(os.Args[2])
		img = glyph(n)
	case len(os.Args) == 4 && (os.Args[1] == "app" || os.Args[1] == "app-dark"):
		n, _ := strconv.Atoi(os.Args[2])
		img = app(n, os.Args[1] == "app-dark")
	default:
		fmt.Fprintln(os.Stderr, "usage: gen.go tray <out.png> | gen.go app|app-dark <size> <out.png>")
		os.Exit(2)
	}
	f, err := os.Create(os.Args[len(os.Args)-1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// trayK is the tray bird's size in its 22pt square: drawn to the edges it
// was a full 22pt wide, bigger than the menu bar's own glyphs (@SherlockYoYo:
// 感觉这只鸟有些大); at .8 it is about 18pt by 14pt, centred.
const trayK = 0.8

// tray draws the silhouette as a macOS template image (black + alpha).
func tray() image.Image {
	const S = 44
	off := S / 2 * (1 - trayK)
	cov := coverage(parse(smallSVG), S, 8, trayK, off, off)
	img := image.NewNRGBA(image.Rect(0, 0, S, S))
	for i, a := range cov {
		img.SetNRGBA(i%S, i/S, color.NRGBA{0, 0, 0, uint8(a*255 + 0.5)})
	}
	return img
}

// The tray mark's motion is the header logo's: the whole mark bobbing about
// its base through the same keyframes over the same .9s,
// cubic-bezier(.3, 0, .3, 1) between each two of them.
const (
	flapSecs  = 0.9
	flapEvery = 0.03 // a frame every 30ms
)

// A keyframe: at is the fraction of the animation, deg the rotation.
type key struct{ at, deg float64 }

var (
	bobKeys = []key{{0, 0}, {.18, -4}, {.40, 2}, {.62, -1}, {1, 0}}
	// transform-origin, in the viewBox
	bobAt = [2]float64{22, 33}
)

// trayFlap writes the frames the tray icon plays when it is clicked,
// flap-01.png on, one every flapEvery of the animation, the still bird at
// both ends left out.
func trayFlap(dir string, S int) error {
	mark := parse(smallSVG)
	n := int(math.Round(flapSecs / flapEvery))
	const ss = 8
	k := float64(S) / 44
	for f := 1; f < n; f++ {
		t := float64(f) / float64(n)
		bob := keyed(bobKeys, t)
		moved := make([][][2]float64, len(mark))
		for i, r := range mark {
			moved[i] = make([][2]float64, len(r))
			for j, p := range r {
				q := turn(p, bobAt, bob)
				// shrunk about the middle, as tray() draws the still mark
				moved[i][j] = [2]float64{22 + (q[0]-22)*trayK, 22 + (q[1]-22)*trayK}
			}
		}
		hit := mask(moved, S*ss, k*ss)
		img := image.NewNRGBA(image.Rect(0, 0, S, S))
		for py := 0; py < S; py++ {
			for px := 0; px < S; px++ {
				on := 0
				for sy := 0; sy < ss; sy++ {
					for sx := 0; sx < ss; sx++ {
						if hit[(py*ss+sy)*S*ss+px*ss+sx] {
							on++
						}
					}
				}
				img.SetNRGBA(px, py, color.NRGBA{0, 0, 0, uint8(float64(on)/float64(ss*ss)*255 + 0.5)})
			}
		}
		out, err := os.Create(fmt.Sprintf("%s/flap-%02d.png", dir, f))
		if err != nil {
			return err
		}
		err = png.Encode(out, img)
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// keyed is the rotation at t of keyframes played as CSS plays them, each
// step eased by cubic-bezier(.3, 0, .3, 1).
func keyed(ks []key, t float64) float64 {
	for i := 1; i < len(ks); i++ {
		if t <= ks[i].at {
			a, b := ks[i-1], ks[i]
			return a.deg + (b.deg-a.deg)*ease((t-a.at)/(b.at-a.at))
		}
	}
	return ks[len(ks)-1].deg
}

// ease is cubic-bezier(.3, 0, .3, 1): x found for the curve's parameter by
// bisection, and its y.
func ease(x float64) float64 {
	bez := func(u, p1, p2 float64) float64 { return 3*u*(1-u)*(1-u)*p1 + 3*u*u*(1-u)*p2 + u*u*u }
	lo, hi := 0.0, 1.0
	for range 40 {
		m := (lo + hi) / 2
		if bez(m, .3, .3) < x {
			lo = m
		} else {
			hi = m
		}
	}
	return bez((lo+hi)/2, 0, 1)
}

// turn rotates p about o by deg degrees, clockwise on screen as CSS's
// rotate() is.
func turn(p, o [2]float64, deg float64) [2]float64 {
	a := deg * math.Pi / 180
	s, c := math.Sin(a), math.Cos(a)
	dx, dy := p[0]-o[0], p[1]-o[1]
	return [2]float64{o[0] + dx*c - dy*s, o[1] + dx*s + dy*c}
}

// mask is which of an n×n grid of samples, k to a unit, fall in the
// rings, even-odd.
func mask(rings [][][2]float64, n int, k float64) []bool {
	out := make([]bool, n*n)
	var xs []float64
	for y := 0; y < n; y++ {
		v := (float64(y) + 0.5) / k
		xs = xs[:0]
		for _, r := range rings {
			for i, j := 0, len(r)-1; i < len(r); j, i = i, i+1 {
				a, b := r[j], r[i]
				if (a[1] > v) != (b[1] > v) {
					xs = append(xs, (a[0]+(v-a[1])*(b[0]-a[0])/(b[1]-a[1]))*k)
				}
			}
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			lo := int(math.Max(0, math.Ceil(xs[i]-0.5)))
			hi := int(math.Min(float64(n), math.Ceil(xs[i+1]-0.5)))
			for x := lo; x < hi; x++ {
				out[y*n+x] = true
			}
		}
	}
	return out
}

// glyph is the bird alone, black on clear, as big on an n-pixel canvas as
// app draws it on its tile: the layer of build/darwin/AppIcon.icon, which
// macOS 26 lays on a tile of its own, light or dark (#117).
func glyph(n int) image.Image {
	S := float64(n)
	k := S * 0.8 / 44
	o := S/2 - 22*k
	cov := coverage(parse(fullSVG), n, 4, k, o, o)
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	for i, a := range cov {
		img.SetNRGBA(i%n, i/n, color.NRGBA{0, 0, 0, uint8(a*255 + 0.5)})
	}
	return img
}

// app draws a rounded square, white shading to the palest grey, with the
// mark on it in ink.
func app(n int, dark bool) image.Image {
	S := float64(n)
	// macOS icon grid: the squircle occupies ~80% of the canvas
	inset := S * 0.1
	side := S - 2*inset
	radius := side * 0.225
	k := side * 0.8 / 44
	o := inset + side/2 - 22*k
	ss := 4
	if n <= 64 {
		ss = 8
	}
	// under 64px the eye is a grey speck; leave it out
	shape := fullSVG
	if n < 64 {
		shape = smallSVG
	}
	mark := coverage(parse(shape), n, ss, k, o, o)
	bg1, bg2 := 0xff, 0xec // top, bottom
	ink := 0x16
	if dark {
		// the Dock's dark appearance (macOS 26): the mark in white on a
		// dark tile, as macOS would otherwise darken the tile under a black
		// mark and lose it (#117)
		bg1, bg2, ink = 0x3a, 0x22, 0xf4
	}
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var tile float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					fx := float64(x) + (float64(sx)+0.5)/float64(ss)
					fy := float64(y) + (float64(sy)+0.5)/float64(ss)
					if roundRect(fx, fy, inset, inset, side, side, radius) {
						tile++
					}
				}
			}
			if tile == 0 {
				continue
			}
			t := float64(y) / S
			bg := float64(bg1)*(1-t) + float64(bg2)*t
			a := mark[y*n+x]
			g := uint8(bg*(1-a) + float64(ink)*a + 0.5)
			img.SetNRGBA(x, y, color.NRGBA{g, g, g, uint8(tile / float64(ss*ss) * 255)})
		}
	}
	return img
}

// coverage rasterises a shape, scaled by k and moved to (ox, oy), onto an n×n
// grid of pixels with ss×ss samples each, and gives each pixel's coverage in
// 0..1. It fills even-odd, a scanline at a time.
func coverage(rings [][][2]float64, n, ss int, k, ox, oy float64) []float64 {
	cov := make([]float64, n*n)
	per := 1 / float64(ss*ss)
	var xs []float64
	for y := 0; y < n*ss; y++ {
		v := ((float64(y)+0.5)/float64(ss) - oy) / k
		xs = xs[:0]
		for _, r := range rings {
			for i, j := 0, len(r)-1; i < len(r); j, i = i, i+1 {
				a, b := r[j], r[i]
				if (a[1] > v) != (b[1] > v) {
					u := a[0] + (v-a[1])*(b[0]-a[0])/(b[1]-a[1])
					xs = append(xs, (u*k+ox)*float64(ss))
				}
			}
		}
		sort.Float64s(xs)
		row := cov[(y/ss)*n:][:n]
		for i := 0; i+1 < len(xs); i += 2 {
			// sample centres x+0.5 that fall between the two crossings
			lo := int(math.Max(0, math.Ceil(xs[i]-0.5)))
			hi := int(math.Min(float64(n*ss), math.Ceil(xs[i+1]-0.5)))
			for x := lo; x < hi; x++ {
				row[x/ss] += per
			}
		}
	}
	return cov
}

// parse reads the path out of one of the SVGs: absolute M, L, C and Z
// commands, each curve cut into short segments, one polygon per subpath.
func parse(svg string) [][][2]float64 {
	d := svg[strings.Index(svg, ` d="`)+4:]
	d = d[:strings.IndexByte(d, '"')]
	f := strings.Fields(strings.NewReplacer("M", " M ", "L", " L ", "C", " C ", "Z", " Z ").Replace(d))
	num := func(i int) float64 { x, _ := strconv.ParseFloat(f[i], 64); return x }
	var rings [][][2]float64
	var pts [][2]float64
	cmd := ""
	for i := 0; i < len(f); {
		if c := f[i]; c == "M" || c == "L" || c == "C" || c == "Z" {
			cmd = c
			i++
		}
		switch cmd {
		case "M", "L":
			if cmd == "M" && len(pts) > 0 {
				rings, pts = append(rings, pts), nil
			}
			pts = append(pts, [2]float64{num(i), num(i + 1)})
			i += 2
			cmd = "L" // further pairs after M are lines
		case "C":
			p := pts[len(pts)-1]
			c := [8]float64{p[0], p[1], num(i), num(i + 1), num(i + 2), num(i + 3), num(i + 4), num(i + 5)}
			for s := 1; s <= 8; s++ {
				t := float64(s) / 8
				m := 1 - t
				pts = append(pts, [2]float64{
					m*m*m*c[0] + 3*m*m*t*c[2] + 3*m*t*t*c[4] + t*t*t*c[6],
					m*m*m*c[1] + 3*m*m*t*c[3] + 3*m*t*t*c[5] + t*t*t*c[7],
				})
			}
			i += 6
		case "Z":
			if len(pts) > 0 {
				rings, pts = append(rings, pts), nil
			}
		}
	}
	if len(pts) > 0 {
		rings = append(rings, pts)
	}
	return rings
}

// roundRect is a rounded square: true inside it.
func roundRect(px, py, x, y, w, h, r float64) bool {
	if px < x || py < y || px > x+w || py > y+h {
		return false
	}
	cx := math.Max(x+r, math.Min(px, x+w-r))
	cy := math.Max(y+r, math.Min(py, y+h-r))
	return math.Hypot(px-cx, py-cy) <= r
}
