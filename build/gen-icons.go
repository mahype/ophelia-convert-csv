//go:build ignore

// gen-icons writes a small tray icon as both PNG and ICO.
// Run with: go run build/gen-icons.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const size = 32

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{R: 0x1f, G: 0x6f, B: 0xb5, A: 0xff}
	fg := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	// filled rounded-ish square background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if (x < 2 || x >= size-2) && (y < 2 || y >= size-2) {
				continue
			}
			img.Set(x, y, bg)
		}
	}

	// draw a stylised "CSV" — three horizontal bars with a gap
	for _, row := range []int{9, 15, 21} {
		for x := 7; x < size-7; x++ {
			for dy := 0; dy < 3; dy++ {
				img.Set(x, row+dy, fg)
			}
		}
	}

	writePNG(img, "cmd/csv-watcher/icon.png")
	writeICO(img, "cmd/csv-watcher/icon.ico")

	fmt.Println("wrote cmd/csv-watcher/icon.png and cmd/csv-watcher/icon.ico")
}

func writePNG(img image.Image, relPath string) {
	path, _ := filepath.Abs(relPath)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

// writeICO wraps the PNG into a minimal ICO container.
func writeICO(img image.Image, relPath string) {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		panic(err)
	}
	pngBytes := pngBuf.Bytes()

	path, _ := filepath.Abs(relPath)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// ICONDIR header (6 bytes)
	header := struct {
		Reserved uint16
		Type     uint16
		Count    uint16
	}{Reserved: 0, Type: 1, Count: 1}
	_ = binary.Write(f, binary.LittleEndian, header)

	// ICONDIRENTRY (16 bytes)
	const headerSize = 6 + 16
	entry := struct {
		Width    uint8
		Height   uint8
		Colors   uint8
		Reserved uint8
		Planes   uint16
		BitCount uint16
		Size     uint32
		Offset   uint32
	}{
		Width:    size,
		Height:   size,
		Colors:   0,
		Reserved: 0,
		Planes:   1,
		BitCount: 32,
		Size:     uint32(len(pngBytes)),
		Offset:   headerSize,
	}
	_ = binary.Write(f, binary.LittleEndian, entry)

	if _, err := f.Write(pngBytes); err != nil {
		panic(err)
	}
}
