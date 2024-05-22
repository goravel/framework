package color

import (
	"bytes"
	"io"

	"github.com/pterm/pterm"

	"github.com/goravel/framework/contracts/support"
)

const (
	FgBlack Color = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
	// FgDefault revert default FG.
	FgDefault Color = 39
)

// Extra foreground color 90 - 97.
const (
	FgDarkGray Color = iota + 90
	FgLightRed
	FgLightGreen
	FgLightYellow
	FgLightBlue
	FgLightMagenta
	FgLightCyan
	FgLightWhite
	// FgGray is an alias of FgDarkGray.
	FgGray Color = 90
)

var (
	info    = pterm.Info
	warn    = pterm.Warning
	err     = pterm.Error
	debug   = pterm.Debug
	success = pterm.Success
)

// New Functions to create Printer with specific color
func New(color Color) support.Printer {
	return color
}

func Green() support.Printer {
	return New(FgGreen)
}

func Red() support.Printer {
	return New(FgRed)
}

func Blue() support.Printer {
	return New(FgBlue)
}

func Yellow() support.Printer {
	return New(FgYellow)
}

func Cyan() support.Printer {
	return New(FgCyan)
}

func White() support.Printer {
	return New(FgWhite)
}

func Gray() support.Printer {
	return New(FgGray)
}

func Default() support.Printer {
	return New(FgDefault)
}

func Black() support.Printer {
	return New(FgBlack)
}

func Magenta() support.Printer {
	return New(FgMagenta)
}

type Color uint8

func (c Color) Sprint(a ...interface{}) string {
	return pterm.Color(c).Sprint(a...)
}

func (c Color) Sprintln(a ...interface{}) string {
	return pterm.Color(c).Sprintln(a...)
}

func (c Color) Sprintf(format string, a ...interface{}) string {
	return pterm.Color(c).Sprintf(format, a...)
}

func (c Color) Sprintfln(format string, a ...interface{}) string {
	return pterm.Color(c).Sprintfln(format, a...)
}

func (c Color) Print(a ...any) *support.Printer {
	pterm.Color(c).Print(a...)
	p := support.Printer(c)
	return &p
}

func (c Color) Println(a ...any) *support.Printer {
	pterm.Color(c).Println(a...)
	p := support.Printer(c)
	return &p
}

func (c Color) Printf(format string, a ...any) *support.Printer {
	pterm.Color(c).Printf(format, a...)
	p := support.Printer(c)
	return &p
}

func (c Color) Printfln(format string, a ...any) *support.Printer {
	pterm.Color(c).Printfln(format, a...)
	p := support.Printer(c)
	return &p
}

// Quick use color print message

func Debugf(format string, a ...any) { debug.Printf(format, a...) }

func Debugln(a ...any) { debug.Println(a...) }

func Errorf(format string, a ...any) { err.Printf(format, a...) }

func Errorln(a ...any) { err.Println(a...) }

func Infof(format string, a ...any) { info.Printf(format, a...) }

func Infoln(a ...any) { info.Println(a...) }

func Successf(format string, a ...any) { success.Printf(format, a...) }

func Successln(a ...any) { success.Println(a...) }

func Warnf(format string, a ...any) { warn.Printf(format, a...) }

func Warnln(a ...any) { warn.Println(a...) }

// CaptureOutput simulates capturing of os.stdout with a buffer and returns what was written to the screen
func CaptureOutput(f func(w io.Writer)) string {
	var outBuf bytes.Buffer
	pterm.SetDefaultOutput(&outBuf)
	f(&outBuf)

	content := outBuf.String()
	outBuf.Reset()
	return content
}
