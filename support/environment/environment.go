package environment

import "runtime"

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func IsArm() bool {
	return runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"
}

func IsX86() bool {
	return runtime.GOARCH == "386" || runtime.GOARCH == "amd64"
}

func Is64Bit() bool {
	return runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"
}
