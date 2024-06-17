package foundation

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getCurrentAbsolutePath() string {
	executable, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(executable))

	if isTesting() || isAir() || isDirectlyRun() {
		res, _ = os.Getwd()
	}

	return res
}

// isTesting checks if the application is running in testing mode.
func isTesting() bool {
	for _, arg := range os.Args {
		if strings.Contains(arg, "-test.") {
			return true
		}
	}

	return false
}

// isAir checks if the application is running using Air.
func isAir() bool {
	for _, arg := range os.Args {
		if strings.Contains(arg, "/storage/temp") {
			return true
		}
	}

	return false
}

// isDirectlyRun checks if the application is running using go run.
func isDirectlyRun() bool {
	executable, _ := os.Executable()
	return strings.Contains(filepath.Base(executable), os.TempDir()) ||
		(strings.Contains(executable, "/var/folders") && strings.Contains(executable, "/T/go-build")) // macOS
}
