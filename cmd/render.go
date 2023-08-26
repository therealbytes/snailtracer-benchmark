package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/therealbytes/concrete-snailtrace/snailtracer"
)

const (
	originX  = 0
	originY  = 0
	width    = 1024 / 2
	height   = 768 / 2
	spp      = 2
	filename = "out.png"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	routines := runtime.NumCPU()

	var wg sync.WaitGroup
	var lock sync.Mutex

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			s := snailtracer.NewBenchmarkScene()
			h := height / routines
			w := width
			y0 := i * h

			for y := y0; y < y0+h; y++ {
				fmt.Printf("%d is %d%% done\n", i, (y-y0)*100/h)
				for x := 0; x < w; x++ {
					v := s.TracePixel(x, y, spp)

					lock.Lock()
					img.Set(x, y, v)
					lock.Unlock()
				}
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
