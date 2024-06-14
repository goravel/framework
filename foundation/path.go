package foundation

import (
	"log"
	"os"
	"path/filepath"
)

func getCurrentAbsolutePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))

	return res
}
