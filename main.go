package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

var (
	MAX_ITER int = 100
)

func mandelbrotCalc(c complex64) int {
	z := complex64(0)
	n := 0
	for math.Abs(float64(real(z))) <= 2 && n < MAX_ITER {
		z = z*z + c
		n += 1
	}
	return n
}

func main() {
	width := 1200
	height := 800

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	RE_START := -2.
	RE_END := 1.
	IM_START := -1.
	IM_END := 1.

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			complNum := complex(RE_START+float64(x)/float64(width)*(RE_END-RE_START),
				IM_START+float64(y)/float64(height)*(IM_END-IM_START))
			val := mandelbrotCalc(complex64(complNum))
			val = 255 - int(val*255/MAX_ITER)
			col := color.RGBA{uint8(val), uint8(val), uint8(val), 254}
			img.Set(x, y, col)
			// fmt.Printf("x=%v  y=%v  val=%v\n", x, y, val)
		}
	}

	// Encode as PNG.
	f, ok := os.Create(fmt.Sprintf("mandelbrot_%d_%d_%d.png", width, height, MAX_ITER))
	if ok == nil {
		fmt.Println("couldn't create image")
	}
	png.Encode(f, img)
}
