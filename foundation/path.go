package foundation

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func getCurrentAbsolutePath() string {
	dir := getCurrentAbsolutePathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbsolutePathByCaller()
	}

	return dir
}

func getCurrentAbsolutePathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))

	return res
}

func getCurrentAbsolutePathByCaller() string {
	var abPath string
	for i := 0; i < 15; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if ok && strings.HasSuffix(filename, "main.go") {
			abPath = filepath.Dir(filename)
			break
		}
	}

	return abPath
}
