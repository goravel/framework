package foundation

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}

	return dir
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))

	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	for i := 0; i < 15; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if ok && strings.HasSuffix(filename, "main.go") {
			abPath = path.Dir(filename)
			break
		}
	}

	return abPath
}
