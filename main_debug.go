//go:build debug

package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/tychonis/cyanotype/cmd"
)

func main() {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatalf("could not create cpu profile: %v", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatalf("could not start cpu profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	cmd.Run()
}
