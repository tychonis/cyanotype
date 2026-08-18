package catalog

import (
	"fmt"
	"os"
)

func Initialize() error {
	bpcDir := ".bpc"
	stat, err := os.Stat(bpcDir)
	if err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("invalid .bpc format")
		}
		return fmt.Errorf("cyanotype repo already initialized")
	}

	return os.Mkdir(bpcDir, 0755)
}
