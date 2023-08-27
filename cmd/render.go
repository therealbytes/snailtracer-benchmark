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
	done   chan int
}

func (w *worker) renderLine(y int) {
	fmt.Println("Starting worker", w.id, "rendering line", y)
Loop:
	for x := originX; x < originX+width; x++ {
		select {
		case <-w.ctx.Done():
			break Loop
		default:
			v := w.scene.TracePixel(x, y, spp)
			w.canvas.set(x, y, v)
		}
	}
	w.done <- w.id
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
	if routines > 1 {
		routines--
	}
	if routines > height {
		routines = height
	}
	// routines = 1

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived an interrupt, stopping render...")
		cancel()
	}()

	doneChan := make(chan int, routines)
	imgCanvas := &canvas{img: img}

	workers := make([]*worker, routines)
	for i := 0; i < routines; i++ {
		workers[i] = &worker{
			ctx:    ctx,
			id:     i,
			scene:  snailtracer.NewBenchmarkScene(),
			canvas: imgCanvas,
			done:   doneChan,
		}
		go workers[i].renderLine(originY + i)
	}

	var wg sync.WaitGroup
	wg.Add(height)

	nextLine := routines
	linesRendered := 0
	startTime := time.Now()

Loop:
	for id := range doneChan {
		select {
		case <-ctx.Done():
			break Loop
		default:
			wg.Done()
			linesRendered++
			expectedTimeLeft := time.Since(startTime) / time.Duration(linesRendered) * time.Duration(height-linesRendered)
			fmt.Println(linesRendered*100/height, "% done -- Expected time left:", expectedTimeLeft.String())

			if nextLine < height {
				go workers[id].renderLine(originY + nextLine)
				nextLine++
			} else if linesRendered == height {
				break Loop
			}
		}
	}

	select {
	case <-ctx.Done():
	default:
		wg.Wait()
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
