package main

import (
	"log/slog"

	"github.com/tychonis/cyanotype/cmd"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	cmd.Run()
}
