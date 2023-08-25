package main

import (
	"image/png"
	"log"
	"os"

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
	s := snailtracer.NewBenchmarkScene()
	img := s.TraceArea(originX, originY, width, height, spp)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create: %s", err)
	}
	defer file.Close()
	png.Encode(file, img)
}
