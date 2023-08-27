package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/therealbytes/concrete-snailtrace/snailtracer"
)

const (
	originX  = 0
	originY  = 0
	width    = 1024
	height   = 768
	spp      = 1
	filename = "out.png"
)

type worker struct {
	ctx    context.Context
	id     int
	scene  *snailtracer.Scene
	canvas *canvas
	lines  chan int
	done   chan int
}

func (w *worker) render() {
	for y := range w.lines {
		fmt.Println("Starting worker", w.id, "rendering line", y)
		for x := originX; x < originX+width; x++ {
			select {
			case <-w.ctx.Done():
				w.done <- w.id
				return
			default:
				v := w.scene.TracePixel(x, y, spp)
				w.canvas.set(x, y, v)
			}
		}
		w.done <- w.id
	}
}

type canvas struct {
	lock sync.Mutex
	img  *image.RGBA
}

func (c *canvas) set(x, y int, v color.Color) {
	c.lock.Lock()
	c.img.Set(x, height-(originY+y)-1, v)
	c.lock.Unlock()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	routines := runtime.NumCPU()
	if routines > height {
		routines = height
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived an interrupt, stopping render...")
		cancel()
	}()

	lineChan := make(chan int, height)
	doneChan := make(chan int, routines)
	imgCanvas := &canvas{img: img}

	for i := 0; i < height; i++ {
		lineChan <- (originY + i)
	}
	close(lineChan)

	for i := 0; i < routines; i++ {
		w := &worker{
			ctx:    ctx,
			id:     i,
			scene:  snailtracer.NewBenchmarkScene(),
			canvas: imgCanvas,
			lines:  lineChan,
			done:   doneChan,
		}
		go w.render()
	}

	linesRendered := 0
	startTime := time.Now()

	for range doneChan {
		linesRendered++
		expectedTimeLeft := time.Since(startTime) / time.Duration(linesRendered) * time.Duration(height-linesRendered)
		fmt.Println(linesRendered*100/height, "% done -- Expected time left:", expectedTimeLeft.String())
		if linesRendered == height {
			break
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create: %s", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatalf("failed to encode: %s", err)
	}
}
