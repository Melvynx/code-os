package server

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestNeutralizeDarkSVGBackgroundKeepsBrandColor(t *testing.T) {
	t.Parallel()
	svg := []byte(`<svg viewBox="0 0 91 91"><rect width="91" height="91" rx="18" fill="#FF3300"/><path fill="#fff" d="M1"/></svg>`)
	got := neutralizeDarkSVGBackground(svg)
	if !bytes.Contains(got, []byte(`fill="#FF3300"`)) {
		t.Fatalf("stripped brand fill: %s", got)
	}
}

func TestNeutralizeDarkSVGBackgroundRemovesBlackPlate(t *testing.T) {
	t.Parallel()
	svg := []byte(`<svg viewBox="0 0 32 32"><rect width="32" height="32" fill="#111111"/><path fill="#fafafa" d="M8 8h16v16H8z"/></svg>`)
	got := neutralizeDarkSVGBackground(svg)
	if !strings.Contains(string(got), `fill="none"`) || strings.Contains(string(got), `fill="#111111"`) {
		t.Fatalf("dark plate still present: %s", got)
	}
}

func TestExtractLargestICOPNG(t *testing.T) {
	t.Parallel()
	var small, large bytes.Buffer
	if err := png.Encode(&small, solidPNG(8, color.NRGBA{A: 255})); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&large, solidPNG(16, color.NRGBA{R: 10, A: 255})); err != nil {
		t.Fatal(err)
	}
	ico := buildICO(t, [][]byte{small.Bytes(), large.Bytes()})
	payload, ok := extractLargestICOPNG(ico)
	if !ok || !bytes.Equal(payload, large.Bytes()) {
		t.Fatalf("extracted %d bytes ok=%v", len(payload), ok)
	}
}

func TestPunchDarkPNGBackgroundRemovesNavyPlate(t *testing.T) {
	t.Parallel()
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	navy := color.NRGBA{R: 0, G: 8, B: 60, A: 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, navy)
		}
	}
	for y := 5; y < 11; y++ {
		for x := 5; x < 11; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 80, G: 160, B: 255, A: 255})
		}
	}
	var source bytes.Buffer
	if err := png.Encode(&source, img); err != nil {
		t.Fatal(err)
	}
	prepared, err := punchDarkPNGBackground(source.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out, err := png.Decode(bytes.NewReader(prepared))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, cornerA := out.At(0, 0).RGBA()
	_, _, _, markA := out.At(8, 8).RGBA()
	if cornerA != 0 || markA == 0 {
		t.Fatalf("cornerA=%d markA=%d", cornerA, markA)
	}
}

func TestPunchDarkPNGBackgroundMakesMarkReadableOnWhite(t *testing.T) {
	t.Parallel()
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, color.NRGBA{A: 255})
		}
	}
	for y := 4; y < 12; y++ {
		for x := 4; x < 12; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 250, G: 250, B: 250, A: 255})
		}
	}
	var source bytes.Buffer
	if err := png.Encode(&source, img); err != nil {
		t.Fatal(err)
	}
	prepared, err := punchDarkPNGBackground(source.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out, err := png.Decode(bytes.NewReader(prepared))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, cornerA := out.At(0, 0).RGBA()
	r, g, b, markA := out.At(8, 8).RGBA()
	if cornerA != 0 {
		t.Fatalf("corner alpha = %d, want transparent", cornerA)
	}
	if markA == 0 || r > 0x4000 || g > 0x4000 || b > 0x4000 {
		t.Fatalf("mark = %d %d %d %d, want a dark opaque pixel", r, g, b, markA)
	}
}

func solidPNG(size int, pixel color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, pixel)
		}
	}
	return img
}

func buildICO(t *testing.T, images [][]byte) []byte {
	t.Helper()
	header := []byte{0, 0, 1, 0, byte(len(images)), 0}
	offset := 6 + 16*len(images)
	var entries, payloads []byte
	for _, image := range images {
		cfg, err := png.DecodeConfig(bytes.NewReader(image))
		if err != nil {
			t.Fatal(err)
		}
		entry := make([]byte, 16)
		entry[0] = byte(cfg.Width)
		entry[1] = byte(cfg.Height)
		entry[8] = byte(len(image))
		entry[9] = byte(len(image) >> 8)
		entry[10] = byte(len(image) >> 16)
		entry[11] = byte(len(image) >> 24)
		entry[12] = byte(offset)
		entry[13] = byte(offset >> 8)
		entry[14] = byte(offset >> 16)
		entry[15] = byte(offset >> 24)
		entries = append(entries, entry...)
		payloads = append(payloads, image...)
		offset += len(image)
	}
	return append(append(header, entries...), payloads...)
}
