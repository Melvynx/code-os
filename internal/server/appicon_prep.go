package server

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"regexp"
	"strconv"
	"strings"
)

var (
	svgViewBox    = regexp.MustCompile(`(?i)viewBox="([^"]+)"`)
	svgCanvasRect = regexp.MustCompile(`(?i)<rect\b[^>]*>`)
	svgFill       = regexp.MustCompile(`(?i)\bfill="([^"]+)"`)
	svgDimension  = regexp.MustCompile(`(?i)\b(width|height)="([^"]+)"`)
	svgDarkFill   = regexp.MustCompile(`(?i)^#(?:0{3,8}|111|111111|0a0a0a|1a1a1a|000000ff)$|^black$|^rgb\(\s*0\s*,\s*0\s*,\s*0\s*\)$`)
)

func prepareApplicationIcon(body []byte, contentType string) ([]byte, string) {
	if pngBody, ok := extractLargestICOPNG(body); ok {
		body = pngBody
		contentType = "image/png"
	}
	switch {
	case strings.Contains(contentType, "svg") || bytes.Contains(bytes.ToLower(body[:min(len(body), 200)]), []byte("<svg")):
		return neutralizeDarkSVGBackground(body), "image/svg+xml"
	case strings.Contains(contentType, "png") || looksLikePNG(body):
		if prepared, err := punchDarkPNGBackground(body); err == nil {
			return prepared, "image/png"
		}
	}
	return body, contentType
}

func extractLargestICOPNG(body []byte) ([]byte, bool) {
	if len(body) < 6 || body[0] != 0 || body[1] != 0 || body[2] != 1 || body[3] != 0 {
		return nil, false
	}
	count := int(body[4]) | int(body[5])<<8
	best, bestArea := []byte(nil), -1
	for index := 0; index < count; index++ {
		entry := 6 + index*16
		if entry+16 > len(body) {
			return nil, false
		}
		width := int(body[entry])
		height := int(body[entry+1])
		if width == 0 {
			width = 256
		}
		if height == 0 {
			height = 256
		}
		size := int(body[entry+8]) | int(body[entry+9])<<8 | int(body[entry+10])<<16 | int(body[entry+11])<<24
		offset := int(body[entry+12]) | int(body[entry+13])<<8 | int(body[entry+14])<<16 | int(body[entry+15])<<24
		if offset < 0 || size < 8 || offset+size > len(body) {
			continue
		}
		payload := body[offset : offset+size]
		if !looksLikePNG(payload) || width*height < bestArea {
			continue
		}
		best, bestArea = payload, width*height
	}
	return best, len(best) > 0
}

func neutralizeDarkSVGBackground(svg []byte) []byte {
	canvasWidth, canvasHeight := svgCanvasSize(svg)
	updated := svgCanvasRect.ReplaceAllFunc(svg, func(tag []byte) []byte {
		fillMatch := svgFill.FindSubmatch(tag)
		if len(fillMatch) < 2 || !svgDarkFill.Match(bytes.TrimSpace(fillMatch[1])) {
			return tag
		}
		if !svgRectCoversCanvas(tag, canvasWidth, canvasHeight) {
			return tag
		}
		replaced := svgFill.ReplaceAll(tag, []byte(`fill="none"`))
		if bytes.Equal(replaced, tag) {
			return tag
		}
		return replaced
	})
	return updated
}

func svgCanvasSize(svg []byte) (float64, float64) {
	if view := svgViewBox.FindSubmatch(svg); len(view) >= 2 {
		parts := strings.Fields(string(view[1]))
		if len(parts) == 4 {
			width, _ := strconv.ParseFloat(parts[2], 64)
			height, _ := strconv.ParseFloat(parts[3], 64)
			if width > 0 && height > 0 {
				return width, height
			}
		}
	}
	var width, height float64
	for _, match := range svgDimension.FindAllSubmatch(svg, 8) {
		if len(match) < 3 {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSuffix(string(match[2]), "px"), 64)
		if err != nil {
			continue
		}
		if bytes.EqualFold(match[1], []byte("width")) && width == 0 {
			width = value
		}
		if bytes.EqualFold(match[1], []byte("height")) && height == 0 {
			height = value
		}
	}
	return width, height
}

func svgRectCoversCanvas(tag []byte, canvasWidth, canvasHeight float64) bool {
	width, height := 0.0, 0.0
	for _, match := range svgDimension.FindAllSubmatch(tag, 4) {
		if len(match) < 3 {
			continue
		}
		raw := strings.TrimSpace(string(match[2]))
		if raw == "100%" {
			if bytes.EqualFold(match[1], []byte("width")) {
				width = canvasWidth
			} else {
				height = canvasHeight
			}
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSuffix(raw, "px"), 64)
		if err != nil {
			continue
		}
		if bytes.EqualFold(match[1], []byte("width")) {
			width = value
		} else {
			height = value
		}
	}
	if width == 0 || height == 0 {
		return canvasWidth == 0
	}
	if canvasWidth == 0 {
		return width >= 16 && height >= 16 && width == height
	}
	return width >= canvasWidth*0.9 && height >= canvasHeight*0.9
}

func punchDarkPNGBackground(body []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	plate, hasPlate := darkCornerPlate(img)
	out := image.NewNRGBA(bounds)
	var dark, lit, opaque int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := nrgbaAt(img, x, y)
			if pixel.A < 16 || (hasPlate && nearColor(pixel, plate, 28) && luminance(pixel) < 55) {
				if pixel.A >= 16 {
					opaque++
					dark++
				}
				out.SetNRGBA(x, y, color.NRGBA{})
				continue
			}
			if pixel.A >= 16 {
				opaque++
				if pixel.R < 28 && pixel.G < 28 && pixel.B < 28 {
					dark++
					out.SetNRGBA(x, y, color.NRGBA{})
					continue
				}
				if pixel.R > 210 && pixel.G > 210 && pixel.B > 210 {
					lit++
				}
			}
			out.SetNRGBA(x, y, pixel)
		}
	}
	if opaque == 0 || float64(dark)/float64(opaque) < 0.18 {
		return body, nil
	}
	remaining := opaque - dark
	if remaining > 0 && float64(lit)/float64(remaining) > 0.4 {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				pixel := out.NRGBAAt(x, y)
				if pixel.A < 16 || pixel.R < 180 || pixel.G < 180 || pixel.B < 180 {
					continue
				}
				out.SetNRGBA(x, y, color.NRGBA{R: 23, G: 23, B: 23, A: pixel.A})
			}
		}
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, out); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func darkCornerPlate(img image.Image) (color.NRGBA, bool) {
	bounds := img.Bounds()
	samples := []color.NRGBA{
		nrgbaAt(img, bounds.Min.X, bounds.Min.Y),
		nrgbaAt(img, bounds.Max.X-1, bounds.Min.Y),
		nrgbaAt(img, bounds.Min.X, bounds.Max.Y-1),
		nrgbaAt(img, bounds.Max.X-1, bounds.Max.Y-1),
	}
	for _, sample := range samples {
		if sample.A < 16 || luminance(sample) > 50 {
			return color.NRGBA{}, false
		}
	}
	for _, sample := range samples[1:] {
		if !nearColor(sample, samples[0], 22) {
			return color.NRGBA{}, false
		}
	}
	return samples[0], true
}

func nrgbaAt(img image.Image, x, y int) color.NRGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func nearColor(left, right color.NRGBA, tolerance uint8) bool {
	return absDiff(left.R, right.R) <= tolerance && absDiff(left.G, right.G) <= tolerance && absDiff(left.B, right.B) <= tolerance
}

func luminance(pixel color.NRGBA) float64 {
	return 0.2126*float64(pixel.R) + 0.7152*float64(pixel.G) + 0.0722*float64(pixel.B)
}

func absDiff(left, right uint8) uint8 {
	if left > right {
		return left - right
	}
	return right - left
}
