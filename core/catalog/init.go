package catalog

import (
	"fmt"
	"log/slog"
	"os"
)

func Initialize() {
	bpcDir := ".bpc"
	stat, err := os.Stat(bpcDir)
	if err == nil {
		if !stat.IsDir() {
			slog.Error("invalid .bpc format")
		}
		return
	}

	err = os.Mkdir(bpcDir, 0755)
	if err != nil {
		return
	}

	fmt.Println("Initialized empty cyanotype repo in .bpc/")
}
