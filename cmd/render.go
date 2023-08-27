package main

import (
	"context"
	"fmt"
	"image"
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
	width    = 1024 / 8
	height   = 768 / 8
	spp      = 1
	filename = "out.png"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	routines := runtime.NumCPU()
	if routines > height {
		routines = height
	}
	// routines = 1

	var wg sync.WaitGroup
	var lock sync.Mutex

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived an interrupt, stopping services...")
		cancel()
	}()

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := snailtracer.NewBenchmarkScene()
			w := width
			x0 := originX
			h := height / routines
			y0 := originY + i*h

			if i == routines-1 {
				h = height - i*h
			}

			start := time.Now()

			fmt.Println("Starting routine", i, "at", x0, y0, "with dimensions", w, h)

			for y := y0; y < y0+h; y++ {
				for x := x0; x < x0+w; x++ {
					select {
					case <-ctx.Done():
						return
					default:
						v := s.TracePixel(x, y, spp)
						lock.Lock()
						// fmt.Println("Setting pixel", x, y, "to", v)
						img.Set(x-originX, y-originY, v)
						lock.Unlock()
					}
				}
				timeLeft := time.Since(start) / time.Duration(y-y0+1) * time.Duration(h-(y-y0+1))
				fmt.Printf("%d is %d%% done\n", i, (y+1-y0)*100/h)
				fmt.Printf("Estimated time left: %s\n", timeLeft)
			}
		}(i)
	}

	wg.Wait()

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create: %s", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatalf("failed to encode: %s", err)
	}
}
