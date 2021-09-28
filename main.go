/*
Package parallelBandelbrot implements a parallel rendering of the Mandelbrot set

The Mandelbrot set is defined by the set of complex numbers "c"
for which the complex numbers of the sequence "Zn"
remain bounded in absolute value. The sequence "Zn" is defined by:
Z(0) = 0
Z(n+1) = Z(n)^2 + c

The task of rendering the image is split into regions of the image(WorkRange)
and each task is given to a new goroutine(calculatorWorker) to calculate. When a
calculatorWorker has finished calculating a region a new regoin gets assigned to it.
*/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
)

// Struct for work item
type workRange struct {
	minX, maxX, minY, maxY int
}

type pixel struct {
	x, y int
	col  color.RGBA
}

// Image and concurrency parameters
var (
	MAX_ITER  int = 255
	IMG_SCALE int = 10

	RE_START float64 = -2.
	RE_END   float64 = 1.
	IM_START float64 = -1.
	IM_END   float64 = 1.
	bound    float64 = 2

	width       int = 600 * IMG_SCALE
	height      int = 400 * IMG_SCALE
	numOfPixels     = width * height

	numWorkTasks int = 64
	numThreads   int = 12

	printCalculatingProgress bool = true
	printDrawingProgress     bool = true
	printWorkItems           bool = false
)

func mandelbrotCalc(c complex64) int {
	z := complex64(0)
	n := 0
	for math.Abs(float64(real(z))) <= bound && n < MAX_ITER {
		z = z*z + c
		n += 1
	}
	return n
}
func calculatorWorker(work workRange, pixelBuffer chan pixel, freeThreads chan bool) {
	for x := work.minX; x < work.maxX; x++ {
		for y := work.minY; y < work.maxY; y++ {
			complNum := complex(RE_START+float64(x)/float64(width)*(RE_END-RE_START),
				IM_START+float64(y)/float64(height)*(IM_END-IM_START))
			val := mandelbrotCalc(complex64(complNum))
			val = 255 - int(val*255/MAX_ITER)
			col := color.RGBA{uint8(val), uint8(val), uint8(val), 254}
			pixelBuffer <- pixel{x, y, col}
		}
	}
	freeThreads <- true
}
func drawingWorker(pixelBuffer chan pixel, image *image.RGBA, pixelCount *int, wg *sync.WaitGroup) {
	for p := range pixelBuffer {
		image.Set(p.x, p.y, p.col)
		*pixelCount++
		if printDrawingProgress && *pixelCount%(numOfPixels/10) == 0 {
			fmt.Print("█")
		}
		if *pixelCount == numOfPixels {
			break
		}
	}
	fmt.Println()
	wg.Done()
}

func workCreator(workBuffer chan workRange) {
	r := int(math.Sqrt(float64(numWorkTasks)))
	work_width := width / r
	work_height := height / r

	for i := 0; i < r; i++ {
		for j := 0; j < r; j++ {
			w := workRange{i * work_width, (i + 1) * work_width,
				j * work_height, (j + 1) * work_height}
			if printWorkItems {
				fmt.Println(w)
			}
			workBuffer <- w
		}
	}

}

func calculatorWorkerStarter(workBuffer chan workRange, pixelBuffer chan pixel, freeThreads chan bool) {
	for t := 0; t < numThreads; t++ {
		freeThreads <- true
	}
	workCounter := numWorkTasks
	for w := range workBuffer { // For every work task(workRange) start a goroutine
		<-freeThreads // continue execution when theres a free thread
		go calculatorWorker(w, pixelBuffer, freeThreads)
		workCounter--
		if workCounter == 0 {
			return
		}
		if printCalculatingProgress && workCounter%(workCounter/10+1) == 0 {
			fmt.Print("█")
		}
	}
	if printCalculatingProgress {
		fmt.Println()
	}
}

func main() {
	// Image initialization
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	image := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// How many pixels have been rendered
	pixelCount := 0

	// Initialize channels to pass work items, pixels to-be-rendered and threads who are free to start working
	freeThreads := make(chan bool, numThreads)
	pixelBuffer := make(chan pixel, numOfPixels)
	workBuffer := make(chan workRange, numWorkTasks)

	// Create work tasks to be distributed to threads to calculate
	workCreator(workBuffer)

	// Give work tasks to workers
	go calculatorWorkerStarter(workBuffer, pixelBuffer, freeThreads)

	// Sync group is used to signal when the drawing thread has finished rendering the image
	var wg sync.WaitGroup
	wg.Add(1)
	// Start drawing thread
	go drawingWorker(pixelBuffer, image, &pixelCount, &wg)

	// Wait for drawing thread to render image
	wg.Wait()
	fmt.Println("Pixels rendered:", pixelCount)

	// Create file
	imageName := fmt.Sprintf("mandelbrot_%d_%d_%d.png", width, height, MAX_ITER)
	f, err := os.Create(imageName)
	if err != nil {
		// Print error if creating file failed
		fmt.Println("err:", err)
	} else {
		// Encode file as PNG
		png.Encode(f, image)
		fmt.Println("Image created:", imageName)
	}
}
