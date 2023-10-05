package env

import "runtime"

// IsWindows returns whether the current operating system is Windows.
// IsWindows 返回当前操作系统是否为 Windows。
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux returns whether the current operating system is Linux.
// IsLinux 返回当前操作系统是否为 Linux。
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsDarwin returns whether the current operating system is Darwin.
// IsDarwin 返回当前操作系统是否为 Darwin。
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsArm returns whether the current CPU architecture is ARM.
// IsArm 返回当前 CPU 架构是否为 ARM。
func IsArm() bool {
	return runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"
}

// IsX86 returns whether the current CPU architecture is X86.
// IsX86 返回当前 CPU 架构是否为 X86。
func IsX86() bool {
	return runtime.GOARCH == "386" || runtime.GOARCH == "amd64"
}

// Is64Bit returns whether the current CPU architecture is 64-bit.
// Is64Bit 返回当前 CPU 架构是否为 64 位。
func Is64Bit() bool {
	return runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"
}
